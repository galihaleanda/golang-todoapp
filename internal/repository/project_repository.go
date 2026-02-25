package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type projectRepository struct {
	db *sqlx.DB
}

// NewProjectRepository creates a new PostgreSQL-backed ProjectRepository.
func NewProjectRepository(db *sqlx.DB) domain.ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (id, user_id, name, description, type, color, created_at, updated_at)
		VALUES (:id, :user_id, :name, :description, :type, :color, :created_at, :updated_at)`

	if _, err := r.db.NamedExecContext(ctx, query, project); err != nil {
		return fmt.Errorf("projectRepository.Create: %w", mapDBError(err))
	}
	return nil
}

func (r *projectRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	query := `
		SELECT p.*, COUNT(t.id) AS task_count
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id AND t.deleted_at IS NULL
		WHERE p.id = $1 AND p.deleted_at IS NULL
		GROUP BY p.id`

	if err := r.db.GetContext(ctx, &project, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("projectRepository.FindByID: %w", err)
	}
	return &project, nil
}

func (r *projectRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Project, error) {
	var projects []*domain.Project
	query := `
		SELECT p.*, COUNT(t.id) AS task_count
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id AND t.deleted_at IS NULL
		WHERE p.user_id = $1 AND p.deleted_at IS NULL
		GROUP BY p.id
		ORDER BY p.created_at DESC`

	if err := r.db.SelectContext(ctx, &projects, query, userID); err != nil {
		return nil, fmt.Errorf("projectRepository.ListByUserID: %w", err)
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = :name, description = :description, type = :type, color = :color, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	res, err := r.db.NamedExecContext(ctx, query, project)
	if err != nil {
		return fmt.Errorf("projectRepository.Update: %w", mapDBError(err))
	}
	return checkRowsAffected(res)
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE projects SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("projectRepository.Delete: %w", err)
	}
	return checkRowsAffected(res)
}
