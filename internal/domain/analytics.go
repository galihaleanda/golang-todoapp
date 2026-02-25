package domain

import "time"

// DailyStats holds productivity stats for a single day.
type DailyStats struct {
	Date          time.Time `json:"date" db:"date"`
	Completed     int       `json:"completed" db:"completed"`
	Created       int       `json:"created" db:"created"`
	AvgTimeHours  float64   `json:"avg_completion_time_hours" db:"avg_completion_time_hours"`
}

// AnalyticsDashboard aggregates all productivity metrics.
type AnalyticsDashboard struct {
	// Overall
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	CompletionRate float64 `json:"completion_rate_percent"`
	OverdueTasks   int     `json:"overdue_tasks"`

	// This week
	CompletedThisWeek     int     `json:"completed_this_week"`
	AvgCompletionTimeHours float64 `json:"avg_completion_time_hours"`

	// Best day
	MostProductiveDay string `json:"most_productive_day"` // e.g. "Monday"

	// Weekly breakdown (last 7 days)
	WeeklyBreakdown []DailyStats `json:"weekly_breakdown"`

	// Priority breakdown
	HighPriorityPending   int `json:"high_priority_pending"`
	MediumPriorityPending int `json:"medium_priority_pending"`
	LowPriorityPending    int `json:"low_priority_pending"`
}
