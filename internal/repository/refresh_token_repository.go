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

type refreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository creates a new PostgreSQL-backed RefreshTokenRepository.
func NewRefreshTokenRepository(db *sqlx.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, device_id, user_agent, expires_at, created_at)
		VALUES (:id, :user_id, :token, :device_id, :user_agent, :expires_at, :created_at)`

	if _, err := r.db.NamedExecContext(ctx, query, token); err != nil {
		return fmt.Errorf("refreshTokenRepository.Create: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token = $1`
	if err := r.db.GetContext(ctx, &rt, query, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("refreshTokenRepository.FindByToken: %w", err)
	}
	return &rt, nil
}

func (r *refreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("refreshTokenRepository.DeleteByToken: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("refreshTokenRepository.DeleteByUserID: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return fmt.Errorf("refreshTokenRepository.DeleteExpired: %w", err)
	}
	return nil
}
