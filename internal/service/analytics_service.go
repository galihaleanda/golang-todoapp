package service

import (
	"context"
	"fmt"
	"time"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/google/uuid"
)

// AnalyticsService handles analytics use cases.
type AnalyticsService struct {
	analyticsRepo domain.AnalyticsRepository
}

// NewAnalyticsService constructs an AnalyticsService with its dependencies.
func NewAnalyticsService(analyticsRepo domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: analyticsRepo}
}

// GetDashboard returns the full productivity dashboard for a user.
func (s *AnalyticsService) GetDashboard(ctx context.Context, userID uuid.UUID) (*domain.AnalyticsDashboard, error) {
	dash, err := s.analyticsRepo.GetDashboard(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("analyticsService.GetDashboard: %w", err)
	}
	return dash, nil
}

// GetDailyStats returns day-by-day stats for a custom date range.
func (s *AnalyticsService) GetDailyStats(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.DailyStats, error) {
	if from.After(to) {
		return nil, fmt.Errorf("from date must be before to date")
	}
	if to.Sub(from).Hours() > 24*90 { // max 90 days
		return nil, fmt.Errorf("date range must not exceed 90 days")
	}

	stats, err := s.analyticsRepo.GetDailyStats(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("analyticsService.GetDailyStats: %w", err)
	}
	return stats, nil
}
