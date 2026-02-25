package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRepository defines data access for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RefreshTokenRepository defines data access for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// TaskRepository defines data access for tasks.
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id uuid.UUID) (*Task, error)
	List(ctx context.Context, userID uuid.UUID, filter TaskFilter, page, limit int) ([]*Task, int, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	FindOverdue(ctx context.Context, userID uuid.UUID) ([]*Task, error)
}

// ProjectRepository defines data access for projects.
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*Project, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AnalyticsRepository defines data access for analytics queries.
type AnalyticsRepository interface {
	GetDashboard(ctx context.Context, userID uuid.UUID) (*AnalyticsDashboard, error)
	GetDailyStats(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]DailyStats, error)
}
