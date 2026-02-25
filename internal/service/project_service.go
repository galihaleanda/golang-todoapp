package service

import (
	"context"
	"fmt"
	"time"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProjectService handles project management use cases.
type ProjectService struct {
	projectRepo domain.ProjectRepository
	log         *logrus.Logger
}

// NewProjectService constructs a ProjectService with its dependencies.
func NewProjectService(projectRepo domain.ProjectRepository, log *logrus.Logger) *ProjectService {
	return &ProjectService{projectRepo: projectRepo, log: log}
}

// Create creates a new project for the authenticated user.
func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateProjectRequest) (*domain.Project, error) {
	now := time.Now()
	color := req.Color
	if color == "" {
		color = "#6366F1" // default indigo
	}

	project := &domain.Project{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Color:       color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("projectService.Create: %w", err)
	}

	s.log.WithFields(logrus.Fields{"project_id": project.ID, "user_id": userID}).Info("project created")
	return project, nil
}

// GetByID retrieves a project, enforcing ownership.
func (s *ProjectService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return project, nil
}

// List returns all projects for the authenticated user.
func (s *ProjectService) List(ctx context.Context, userID uuid.UUID) ([]*domain.Project, error) {
	projects, err := s.projectRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("projectService.List: %w", err)
	}
	return projects, nil
}

// Update applies partial updates to a project, enforcing ownership.
func (s *ProjectService) Update(ctx context.Context, id, userID uuid.UUID, req *domain.UpdateProjectRequest) (*domain.Project, error) {
	project, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Type != nil {
		project.Type = *req.Type
	}
	if req.Color != nil {
		project.Color = *req.Color
	}

	project.UpdatedAt = time.Now()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("projectService.Update: %w", err)
	}

	return project, nil
}

// Delete soft-deletes a project, enforcing ownership.
func (s *ProjectService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	project, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	if err := s.projectRepo.Delete(ctx, project.ID); err != nil {
		return fmt.Errorf("projectService.Delete: %w", err)
	}

	return nil
}
