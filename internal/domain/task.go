package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// TaskPriority represents the priority level of a task.
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// Task represents the core task entity.
type Task struct {
	ID             uuid.UUID    `json:"id" db:"id"`
	UserID         uuid.UUID    `json:"user_id" db:"user_id"`
	ProjectID      *uuid.UUID   `json:"project_id,omitempty" db:"project_id"`
	Title          string       `json:"title" db:"title"`
	Description    string       `json:"description" db:"description"`
	Status         TaskStatus   `json:"status" db:"status"`
	Priority       TaskPriority `json:"priority" db:"priority"`
	EstimatedHours *float64     `json:"estimated_hours,omitempty" db:"estimated_hours"`
	DueDate        *time.Time   `json:"due_date,omitempty" db:"due_date"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	SmartScore     float64      `json:"smart_score" db:"smart_score"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
}

// IsOverdue returns true when a task has passed its due date and is not done.
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == TaskStatusDone {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// CalculateSmartScore computes a priority score based on multiple factors.
// Higher score = higher urgency.
func (t *Task) CalculateSmartScore() float64 {
	score := 0.0

	// Base score from manual priority
	switch t.Priority {
	case TaskPriorityHigh:
		score += 30
	case TaskPriorityMedium:
		score += 20
	case TaskPriorityLow:
		score += 10
	}

	// Due-date proximity factor (max +50)
	if t.DueDate != nil {
		now := time.Now()
		hoursUntilDue := t.DueDate.Sub(now).Hours()

		switch {
		case hoursUntilDue < 0:
			// Overdue — penalise heavily, each extra day adds 5
			overdueHours := -hoursUntilDue
			score += 50 + (overdueHours/24)*5
		case hoursUntilDue <= 24:
			score += 50
		case hoursUntilDue <= 72:
			score += 40
		case hoursUntilDue <= 168: // 1 week
			score += 25
		case hoursUntilDue <= 720: // 1 month
			score += 10
		}
	}

	// Status factor
	if t.Status == TaskStatusInProgress {
		score += 15
	}

	// Estimation factor — shorter tasks get slight boost to clear quick wins
	if t.EstimatedHours != nil && *t.EstimatedHours <= 1 {
		score += 5
	}

	return score
}

// TaskFilter holds filter criteria for listing tasks.
type TaskFilter struct {
	Status    *TaskStatus  `form:"status"`
	Priority  *TaskPriority `form:"priority"`
	ProjectID *uuid.UUID   `form:"project_id"`
	Overdue   *bool        `form:"overdue"`
	Search    string       `form:"search"`
}

// CreateTaskRequest is the payload for creating a task.
type CreateTaskRequest struct {
	ProjectID      *uuid.UUID   `json:"project_id"`
	Title          string       `json:"title" validate:"required,min=1,max=255"`
	Description    string       `json:"description" validate:"max=5000"`
	Priority       TaskPriority `json:"priority" validate:"required,oneof=low medium high"`
	EstimatedHours *float64     `json:"estimated_hours" validate:"omitempty,min=0,max=999"`
	DueDate        *time.Time   `json:"due_date"`
}

// UpdateTaskRequest is the payload for updating a task.
type UpdateTaskRequest struct {
	ProjectID      *uuid.UUID   `json:"project_id"`
	Title          *string      `json:"title" validate:"omitempty,min=1,max=255"`
	Description    *string      `json:"description" validate:"omitempty,max=5000"`
	Status         *TaskStatus  `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority       *TaskPriority `json:"priority" validate:"omitempty,oneof=low medium high"`
	EstimatedHours *float64     `json:"estimated_hours" validate:"omitempty,min=0,max=999"`
	DueDate        *time.Time   `json:"due_date"`
}
