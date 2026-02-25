package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type analyticsRepository struct {
	db *sqlx.DB
}

// NewAnalyticsRepository creates a new PostgreSQL-backed AnalyticsRepository.
func NewAnalyticsRepository(db *sqlx.DB) domain.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetDashboard(ctx context.Context, userID uuid.UUID) (*domain.AnalyticsDashboard, error) {
	dash := &domain.AnalyticsDashboard{}

	// Total & completed
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'done') AS completed,
			COUNT(*) FILTER (WHERE due_date < NOW() AND status != 'done') AS overdue
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&dash.TotalTasks, &dash.CompletedTasks, &dash.OverdueTasks)
	if err != nil {
		return nil, fmt.Errorf("analyticsRepository.GetDashboard totals: %w", err)
	}

	if dash.TotalTasks > 0 {
		dash.CompletionRate = float64(dash.CompletedTasks) / float64(dash.TotalTasks) * 100
	}

	// This week completions
	weekStart := time.Now().AddDate(0, 0, -7)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND status = 'done' AND completed_at >= $2`, userID, weekStart,
	).Scan(&dash.CompletedThisWeek)
	if err != nil {
		return nil, fmt.Errorf("analyticsRepository.GetDashboard weekly: %w", err)
	}

	// Average completion time (hours)
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 3600), 0)
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL AND status = 'done' AND completed_at IS NOT NULL`, userID,
	).Scan(&dash.AvgCompletionTimeHours)
	if err != nil {
		return nil, fmt.Errorf("analyticsRepository.GetDashboard avg time: %w", err)
	}

	// Most productive day of week
	err = r.db.QueryRowContext(ctx, `
		SELECT TO_CHAR(completed_at, 'Day')
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL AND status = 'done' AND completed_at IS NOT NULL
		GROUP BY TO_CHAR(completed_at, 'Day'), EXTRACT(DOW FROM completed_at)
		ORDER BY COUNT(*) DESC
		LIMIT 1`, userID,
	).Scan(&dash.MostProductiveDay)
	if err != nil {
		// Not fatal — user may have no completed tasks yet
		dash.MostProductiveDay = "N/A"
	}

	// Priority breakdown (pending only)
	err = r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE priority = 'high') AS high,
			COUNT(*) FILTER (WHERE priority = 'medium') AS medium,
			COUNT(*) FILTER (WHERE priority = 'low') AS low
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL AND status != 'done'`, userID,
	).Scan(&dash.HighPriorityPending, &dash.MediumPriorityPending, &dash.LowPriorityPending)
	if err != nil {
		return nil, fmt.Errorf("analyticsRepository.GetDashboard priority: %w", err)
	}

	// Weekly breakdown
	daily, err := r.GetDailyStats(ctx, userID, weekStart, time.Now())
	if err != nil {
		return nil, err
	}
	dash.WeeklyBreakdown = daily

	return dash, nil
}

func (r *analyticsRepository) GetDailyStats(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.DailyStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			DATE(completed_at) AS date,
			COUNT(*) FILTER (WHERE status = 'done') AS completed,
			COUNT(*) FILTER (WHERE DATE(created_at) = DATE(completed_at)) AS created,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 3600) FILTER (WHERE status = 'done'), 0) AS avg_completion_time_hours
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		  AND completed_at BETWEEN $2 AND $3
		GROUP BY DATE(completed_at)
		ORDER BY DATE(completed_at) ASC`, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("analyticsRepository.GetDailyStats: %w", err)
	}
	defer rows.Close()

	var stats []domain.DailyStats
	for rows.Next() {
		var s domain.DailyStats
		if err := rows.Scan(&s.Date, &s.Completed, &s.Created, &s.AvgTimeHours); err != nil {
			return nil, fmt.Errorf("analyticsRepository.GetDailyStats scan: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
