package service

import (
	"context"
	"log"
	"sync"
	"time"

	"task-management-api/internal/config"
	"task-management-api/internal/domain"
	"task-management-api/internal/repository"
)

type WorkerService interface {
	Start(ctx context.Context)
	EnqueueTask(taskID string)
}

type workerService struct {
	taskRepo     repository.TaskRepository
	config       *config.Config
	taskQueue    chan string
	processedIDs sync.Map
	wg           sync.WaitGroup
}

func NewWorkerService(taskRepo repository.TaskRepository, cfg *config.Config) WorkerService {
	return &workerService{
		taskRepo:  taskRepo,
		config:    cfg,
		taskQueue: make(chan string, 100),
	}
}

func (w *workerService) Start(ctx context.Context) {
	// Start worker goroutines
	for i := 0; i < 5; i++ {
		w.wg.Add(1)
		go w.worker(ctx)
	}

	// Start periodic scanner
	w.wg.Add(1)
	go w.scanner(ctx)

	log.Println("Worker service started")
}

func (w *workerService) EnqueueTask(taskID string) {
	select {
	case w.taskQueue <- taskID:
		log.Printf("Task %s enqueued for auto-completion", taskID)
	default:
		log.Printf("Task queue full, skipping task %s", taskID)
	}
}

func (w *workerService) worker(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down")
			return
		case taskID := <-w.taskQueue:
			w.processTask(taskID)
		}
	}
}

func (w *workerService) processTask(taskID string) {
	// Check if already processed
	if _, exists := w.processedIDs.LoadOrStore(taskID, true); exists {
		return
	}

	// Wait for the configured duration
	time.Sleep(time.Duration(w.config.Worker.AutoCompleteMinutes) * time.Minute)

	// Fetch task to verify it still needs auto-completion
	task, err := w.taskRepo.FindByID(taskID)
	if err != nil {
		log.Printf("Error fetching task %s: %v", taskID, err)
		w.processedIDs.Delete(taskID)
		return
	}

	if task == nil {
		log.Printf("Task %s not found (may have been deleted)", taskID)
		w.processedIDs.Delete(taskID)
		return
	}

	// Only auto-complete if still pending or in progress
	if task.Status == domain.StatusPending || task.Status == domain.StatusInProgress {
		if err := w.taskRepo.UpdateStatus(taskID, domain.StatusCompleted); err != nil {
			log.Printf("Error auto-completing task %s: %v", taskID, err)
		} else {
			log.Printf("Task %s auto-completed successfully", taskID)
		}
	} else {
		log.Printf("Task %s already completed, skipping auto-completion", taskID)
	}

	w.processedIDs.Delete(taskID)
}

func (w *workerService) scanner(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Scanner shutting down")
			return
		case <-ticker.C:
			w.scanPendingTasks()
		}
	}
}

func (w *workerService) scanPendingTasks() {
	duration := time.Duration(w.config.Worker.AutoCompleteMinutes) * time.Minute
	tasks, err := w.taskRepo.FindPendingTasksOlderThan(duration)
	if err != nil {
		log.Printf("Error scanning pending tasks: %v", err)
		return
	}

	for _, task := range tasks {
		// Check if task is already being processed
		if _, exists := w.processedIDs.Load(task.ID); !exists {
			if err := w.taskRepo.UpdateStatus(task.ID, domain.StatusCompleted); err != nil {
				log.Printf("Error auto-completing task %s: %v", task.ID, err)
			} else {
				log.Printf("Task %s auto-completed by scanner", task.ID)
			}
		}
	}
}
