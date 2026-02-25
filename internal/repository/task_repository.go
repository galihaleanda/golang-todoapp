package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type taskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository creates a new PostgreSQL-backed TaskRepository.
func NewTaskRepository(db *sqlx.DB) domain.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (
			id, user_id, project_id, title, description,
			status, priority, estimated_hours, due_date,
			completed_at, smart_score, created_at, updated_at
		) VALUES (
			:id, :user_id, :project_id, :title, :description,
			:status, :priority, :estimated_hours, :due_date,
			:completed_at, :smart_score, :created_at, :updated_at
		)`

	if _, err := r.db.NamedExecContext(ctx, query, task); err != nil {
		return fmt.Errorf("taskRepository.Create: %w", mapDBError(err))
	}
	return nil
}

func (r *taskRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	var task domain.Task
	query := `SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &task, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("taskRepository.FindByID: %w", err)
	}
	return &task, nil
}

func (r *taskRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	filter domain.TaskFilter,
	page, limit int,
) ([]*domain.Task, int, error) {
	args := []any{userID}
	conditions := []string{"user_id = $1", "deleted_at IS NULL"}
	argIdx := 2

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *filter.Priority)
		argIdx++
	}
	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, *filter.ProjectID)
		argIdx++
	}
	if filter.Overdue != nil && *filter.Overdue {
		conditions = append(conditions, "due_date < NOW() AND status != 'done'")
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx+1,
		))
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern)
		argIdx += 2
	}

	where := strings.Join(conditions, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", where)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("taskRepository.List count: %w", err)
	}

	// Fetch page
	offset := (page - 1) * limit
	listQuery := fmt.Sprintf(
		"SELECT * FROM tasks WHERE %s ORDER BY smart_score DESC, created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	var tasks []*domain.Task
	if err := r.db.SelectContext(ctx, &tasks, listQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("taskRepository.List select: %w", err)
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks SET
			project_id     = :project_id,
			title          = :title,
			description    = :description,
			status         = :status,
			priority       = :priority,
			estimated_hours = :estimated_hours,
			due_date       = :due_date,
			completed_at   = :completed_at,
			smart_score    = :smart_score,
			updated_at     = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	res, err := r.db.NamedExecContext(ctx, query, task)
	if err != nil {
		return fmt.Errorf("taskRepository.Update: %w", mapDBError(err))
	}
	return checkRowsAffected(res)
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tasks SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("taskRepository.Delete: %w", err)
	}
	return checkRowsAffected(res)
}

func (r *taskRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	)
	if err != nil {
		return 0, fmt.Errorf("taskRepository.CountByUserID: %w", err)
	}
	return count, nil
}

func (r *taskRepository) FindOverdue(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
	var tasks []*domain.Task
	query := `
		SELECT * FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND status != 'done' AND due_date < NOW()
		ORDER BY due_date ASC`

	if err := r.db.SelectContext(ctx, &tasks, query, userID); err != nil {
		return nil, fmt.Errorf("taskRepository.FindOverdue: %w", err)
	}
	return tasks, nil
}
