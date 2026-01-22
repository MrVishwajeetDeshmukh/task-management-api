package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"task-management-api/internal/domain"
)

type TaskRepository interface {
	Create(task *domain.Task) error
	FindByID(id string) (*domain.Task, error)
	FindAll(filter domain.TaskFilter, userID string, isAdmin bool) ([]domain.Task, error)
	Update(task *domain.Task) error
	Delete(id string) error
	FindPendingTasksOlderThan(duration time.Duration) ([]domain.Task, error)
	UpdateStatus(id string, status domain.TaskStatus) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, user_id, title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(
		query,
		task.ID,
		task.UserID,
		task.Title,
		task.Description,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *taskRepository) FindByID(id string) (*domain.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	task := &domain.Task{}
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	return task, nil
}

func (r *taskRepository) FindAll(filter domain.TaskFilter, userID string, isAdmin bool) ([]domain.Task, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	if !isAdmin {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, userID)
		argCount++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *filter.Status)
		argCount++
	}

	query := "SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *taskRepository) Update(task *domain.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(
		query,
		task.Title,
		task.Description,
		task.Status,
		task.UpdatedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (r *taskRepository) Delete(id string) error {
	query := "DELETE FROM tasks WHERE id = $1"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *taskRepository) FindPendingTasksOlderThan(duration time.Duration) ([]domain.Task, error) {
	cutoffTime := time.Now().Add(-duration)
	query := `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE (status = $1 OR status = $2) AND created_at < $3
	`
	rows, err := r.db.Query(query, domain.StatusPending, domain.StatusInProgress, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *taskRepository) UpdateStatus(id string, status domain.TaskStatus) error {
	query := `
		UPDATE tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}
