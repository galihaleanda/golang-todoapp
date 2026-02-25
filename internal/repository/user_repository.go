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

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES (:id, :name, :email, :password_hash, :created_at, :updated_at)`

	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return fmt.Errorf("userRepository.Create: %w", mapDBError(err))
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("userRepository.FindByEmail: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET name = :name, email = :email, password_hash = :password_hash, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	res, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("userRepository.Update: %w", mapDBError(err))
	}
	return checkRowsAffected(res)
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	return checkRowsAffected(res)
}
