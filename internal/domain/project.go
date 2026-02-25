package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProjectType classifies a project by context.
type ProjectType string

const (
	ProjectTypePersonal    ProjectType = "personal"
	ProjectTypeWork        ProjectType = "work"
	ProjectTypeSideProject ProjectType = "side_project"
)

// Project groups related tasks.
type Project struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Type        ProjectType `json:"type" db:"type"`
	Color       string      `json:"color" db:"color"` // hex color e.g. "#3B82F6"
	TaskCount   int         `json:"task_count" db:"task_count"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateProjectRequest is the payload for creating a project.
type CreateProjectRequest struct {
	Name        string      `json:"name" validate:"required,min=1,max=100"`
	Description string      `json:"description" validate:"max=500"`
	Type        ProjectType `json:"type" validate:"required,oneof=personal work side_project"`
	Color       string      `json:"color" validate:"omitempty,hexcolor"`
}

// UpdateProjectRequest is the payload for updating a project.
type UpdateProjectRequest struct {
	Name        *string      `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string      `json:"description" validate:"omitempty,max=500"`
	Type        *ProjectType `json:"type" validate:"omitempty,oneof=personal work side_project"`
	Color       *string      `json:"color" validate:"omitempty,hexcolor"`
}
