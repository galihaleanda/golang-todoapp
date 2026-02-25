package service

import (
	"context"
	"fmt"
	"time"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TaskService handles task management use cases.
type TaskService struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	log         *logrus.Logger
}

// NewTaskService constructs a TaskService with its dependencies.
func NewTaskService(taskRepo domain.TaskRepository, projectRepo domain.ProjectRepository, log *logrus.Logger) *TaskService {
	return &TaskService{taskRepo: taskRepo, projectRepo: projectRepo, log: log}
}

// Create creates a new task for the authenticated user.
func (s *TaskService) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateTaskRequest) (*domain.Task, error) {
	// Validate project ownership if provided
	if req.ProjectID != nil {
		if err := s.assertProjectOwner(ctx, *req.ProjectID, userID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	task := &domain.Task{
		ID:             uuid.New(),
		UserID:         userID,
		ProjectID:      req.ProjectID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         domain.TaskStatusTodo,
		Priority:       req.Priority,
		EstimatedHours: req.EstimatedHours,
		DueDate:        req.DueDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	task.SmartScore = task.CalculateSmartScore()

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("taskService.Create: %w", err)
	}

	s.log.WithFields(logrus.Fields{"task_id": task.ID, "user_id": userID}).Info("task created")
	return task, nil
}

// GetByID retrieves a task, enforcing ownership.
func (s *TaskService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return task, nil
}

// List returns a paginated list of tasks for the authenticated user.
func (s *TaskService) List(ctx context.Context, userID uuid.UUID, filter domain.TaskFilter, page, limit int) ([]*domain.Task, int, error) {
	tasks, total, err := s.taskRepo.List(ctx, userID, filter, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("taskService.List: %w", err)
	}
	return tasks, total, nil
}

// Update applies partial updates to a task, enforcing ownership.
func (s *TaskService) Update(ctx context.Context, id, userID uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	task, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Validate project ownership if changing project
	if req.ProjectID != nil {
		if err := s.assertProjectOwner(ctx, *req.ProjectID, userID); err != nil {
			return nil, err
		}
		task.ProjectID = req.ProjectID
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.EstimatedHours != nil {
		task.EstimatedHours = req.EstimatedHours
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if req.Status != nil && *req.Status != task.Status {
		task.Status = *req.Status
		// Set completed_at when marking as done
		if task.Status == domain.TaskStatusDone {
			now := time.Now()
			task.CompletedAt = &now
		} else {
			task.CompletedAt = nil
		}
	}

	task.SmartScore = task.CalculateSmartScore()
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("taskService.Update: %w", err)
	}

	return task, nil
}

// Delete soft-deletes a task, enforcing ownership.
func (s *TaskService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	task, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if err := s.taskRepo.Delete(ctx, task.ID); err != nil {
		return fmt.Errorf("taskService.Delete: %w", err)
	}

	return nil
}

// RefreshSmartScores recalculates smart scores for all pending user tasks.
// Intended to be called periodically (e.g. via a cron job).
func (s *TaskService) RefreshSmartScores(ctx context.Context, userID uuid.UUID) error {
	pending := domain.TaskStatusTodo
	filter := domain.TaskFilter{Status: &pending}
	tasks, _, err := s.taskRepo.List(ctx, userID, filter, 1, 1000)
	if err != nil {
		return fmt.Errorf("taskService.RefreshSmartScores list: %w", err)
	}

	for _, task := range tasks {
		task.SmartScore = task.CalculateSmartScore()
		task.UpdatedAt = time.Now()
		if err := s.taskRepo.Update(ctx, task); err != nil {
			s.log.WithError(err).WithField("task_id", task.ID).Warn("failed to update smart score")
		}
	}

	return nil
}

func (s *TaskService) assertProjectOwner(ctx context.Context, projectID, userID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project.UserID != userID {
		return domain.ErrForbidden
	}
	return nil
}
