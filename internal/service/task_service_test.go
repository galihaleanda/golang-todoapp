package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/galihaleanda/todo-app/internal/service"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock implementations ---

type mockTaskRepo struct{ mock.Mock }

func (m *mockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	return m.Called(ctx, task).Error(0)
}
func (m *mockTaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}
func (m *mockTaskRepo) List(ctx context.Context, userID uuid.UUID, filter domain.TaskFilter, page, limit int) ([]*domain.Task, int, error) {
	args := m.Called(ctx, userID, filter, page, limit)
	return args.Get(0).([]*domain.Task), args.Int(1), args.Error(2)
}
func (m *mockTaskRepo) Update(ctx context.Context, task *domain.Task) error {
	return m.Called(ctx, task).Error(0)
}
func (m *mockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockTaskRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockTaskRepo) FindOverdue(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Task), args.Error(1)
}

type mockProjectRepo struct{ mock.Mock }

func (m *mockProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}
func (m *mockProjectRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Project, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Project), args.Error(1)
}
func (m *mockProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- Tests ---

func newTaskService(taskRepo domain.TaskRepository, projectRepo domain.ProjectRepository) *service.TaskService {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel) // silence logs in tests
	return service.NewTaskService(taskRepo, projectRepo, log)
}

func TestTaskService_Create_Success(t *testing.T) {
	taskRepo := &mockTaskRepo{}
	projectRepo := &mockProjectRepo{}
	svc := newTaskService(taskRepo, projectRepo)

	userID := uuid.New()
	req := &domain.CreateTaskRequest{
		Title:    "Write tests",
		Priority: domain.TaskPriorityHigh,
	}

	taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

	task, err := svc.Create(context.Background(), userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "Write tests", task.Title)
	assert.Equal(t, domain.TaskStatusTodo, task.Status)
	assert.Equal(t, userID, task.UserID)
	assert.Greater(t, task.SmartScore, 0.0)
	taskRepo.AssertExpectations(t)
}

func TestTaskService_Create_WithProject_NotOwner(t *testing.T) {
	taskRepo := &mockTaskRepo{}
	projectRepo := &mockProjectRepo{}
	svc := newTaskService(taskRepo, projectRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()

	project := &domain.Project{ID: projectID, UserID: otherUserID}
	projectRepo.On("FindByID", mock.Anything, projectID).Return(project, nil)

	req := &domain.CreateTaskRequest{
		Title:     "Task in foreign project",
		Priority:  domain.TaskPriorityLow,
		ProjectID: &projectID,
	}

	_, err := svc.Create(context.Background(), userID, req)

	assert.ErrorIs(t, err, domain.ErrForbidden)
	taskRepo.AssertNotCalled(t, "Create")
}

func TestTaskService_Update_CompletionSetsCompletedAt(t *testing.T) {
	taskRepo := &mockTaskRepo{}
	projectRepo := &mockProjectRepo{}
	svc := newTaskService(taskRepo, projectRepo)

	userID := uuid.New()
	taskID := uuid.New()

	existing := &domain.Task{
		ID:       taskID,
		UserID:   userID,
		Title:    "Pending task",
		Status:   domain.TaskStatusInProgress,
		Priority: domain.TaskPriorityMedium,
	}

	taskRepo.On("FindByID", mock.Anything, taskID).Return(existing, nil)
	taskRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

	done := domain.TaskStatusDone
	req := &domain.UpdateTaskRequest{Status: &done}

	updated, err := svc.Update(context.Background(), taskID, userID, req)

	assert.NoError(t, err)
	assert.Equal(t, domain.TaskStatusDone, updated.Status)
	assert.NotNil(t, updated.CompletedAt)
	assert.WithinDuration(t, time.Now(), *updated.CompletedAt, 5*time.Second)
}

func TestTask_CalculateSmartScore_Overdue(t *testing.T) {
	pastDue := time.Now().Add(-48 * time.Hour) // 2 days overdue
	task := &domain.Task{
		Priority: domain.TaskPriorityHigh,
		DueDate:  &pastDue,
		Status:   domain.TaskStatusTodo,
	}

	score := task.CalculateSmartScore()

	// High priority (30) + overdue base (50) + 2 days * 5 = 90
	assert.GreaterOrEqual(t, score, 80.0, "overdue high priority task should have high score")
}

func TestTask_IsOverdue(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name     string
		task     domain.Task
		expected bool
	}{
		{
			name:     "past due and not done",
			task:     domain.Task{DueDate: &past, Status: domain.TaskStatusTodo},
			expected: true,
		},
		{
			name:     "past due but done",
			task:     domain.Task{DueDate: &past, Status: domain.TaskStatusDone},
			expected: false,
		},
		{
			name:     "future due date",
			task:     domain.Task{DueDate: &future, Status: domain.TaskStatusTodo},
			expected: false,
		},
		{
			name:     "no due date",
			task:     domain.Task{Status: domain.TaskStatusTodo},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.task.IsOverdue())
		})
	}
}
