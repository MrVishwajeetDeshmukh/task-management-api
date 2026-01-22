package service

import (
	"fmt"
	"time"

	"task-management-api/internal/domain"
	"task-management-api/internal/repository"

	"github.com/google/uuid"
)

type TaskService interface {
	Create(req domain.CreateTaskRequest, userID string) (*domain.Task, error)
	GetByID(id, userID string, isAdmin bool) (*domain.Task, error)
	List(filter domain.TaskFilter, userID string, isAdmin bool) ([]domain.Task, error)
	Update(id string, req domain.UpdateTaskRequest, userID string, isAdmin bool) (*domain.Task, error)
	Delete(id, userID string, isAdmin bool) error
}

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{taskRepo: taskRepo}
}

func (s *taskService) Create(req domain.CreateTaskRequest, userID string) (*domain.Task, error) {
	task := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetByID(id, userID string, isAdmin bool) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Authorization check
	if !isAdmin && task.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	return task, nil
}

func (s *taskService) List(filter domain.TaskFilter, userID string, isAdmin bool) ([]domain.Task, error) {
	return s.taskRepo.FindAll(filter, userID, isAdmin)
}

func (s *taskService) Update(id string, req domain.UpdateTaskRequest, userID string, isAdmin bool) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Authorization check
	if !isAdmin && task.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	// Update fields
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, fmt.Errorf("invalid status")
		}
		task.Status = *req.Status
	}
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) Delete(id, userID string, isAdmin bool) error {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	// Authorization check
	if !isAdmin && task.UserID != userID {
		return fmt.Errorf("unauthorized access")
	}

	return s.taskRepo.Delete(id)
}
