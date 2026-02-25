package handler

import (
	"errors"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/galihaleanda/todo-app/internal/middleware"
	"github.com/galihaleanda/todo-app/internal/service"
	"github.com/galihaleanda/todo-app/internal/validator"
	"github.com/galihaleanda/todo-app/pkg/pagination"
	"github.com/galihaleanda/todo-app/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler exposes task CRUD endpoints.
type TaskHandler struct {
	taskSvc *service.TaskService
}

// NewTaskHandler creates a TaskHandler.
func NewTaskHandler(taskSvc *service.TaskService) *TaskHandler {
	return &TaskHandler{taskSvc: taskSvc}
}

// Create godoc
// @Summary Create a task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body domain.CreateTaskRequest true "Task payload"
// @Success 201 {object} response.Envelope{data=domain.Task}
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var req domain.CreateTaskRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	task, err := h.taskSvc.Create(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, task)
}

// List godoc
// @Summary List tasks
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status (todo|in_progress|done)"
// @Param priority query string false "Filter by priority (low|medium|high)"
// @Param project_id query string false "Filter by project UUID"
// @Param overdue query bool false "Show only overdue tasks"
// @Param search query string false "Full-text search"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Envelope{data=[]domain.Task}
// @Router /tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	pag := pagination.FromContext(c)

	filter := domain.TaskFilter{}
	if s := c.Query("status"); s != "" {
		status := domain.TaskStatus(s)
		filter.Status = &status
	}
	if p := c.Query("priority"); p != "" {
		priority := domain.TaskPriority(p)
		filter.Priority = &priority
	}
	if pid := c.Query("project_id"); pid != "" {
		id, err := uuid.Parse(pid)
		if err == nil {
			filter.ProjectID = &id
		}
	}
	if c.Query("overdue") == "true" {
		t := true
		filter.Overdue = &t
	}
	filter.Search = c.Query("search")

	tasks, total, err := h.taskSvc.List(c.Request.Context(), userID, filter, pag.Page, pag.Limit)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, tasks, pag.Page, pag.Limit, total)
}

// GetByID godoc
// @Summary Get a task by ID
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task UUID"
// @Success 200 {object} response.Envelope{data=domain.Task}
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid task id", nil)
		return
	}

	task, err := h.taskSvc.GetByID(c.Request.Context(), id, middleware.CurrentUserID(c))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, task)
}

// Update godoc
// @Summary Update a task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Task UUID"
// @Param body body domain.UpdateTaskRequest true "Update payload"
// @Success 200 {object} response.Envelope{data=domain.Task}
// @Router /tasks/{id} [patch]
func (h *TaskHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid task id", nil)
		return
	}

	var req domain.UpdateTaskRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	task, err := h.taskSvc.Update(c.Request.Context(), id, middleware.CurrentUserID(c), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, task)
}

// Delete godoc
// @Summary Delete a task
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task UUID"
// @Success 200 {object} response.Envelope
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid task id", nil)
		return
	}

	if err := h.taskSvc.Delete(c.Request.Context(), id, middleware.CurrentUserID(c)); err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "task deleted"})
}

func (h *TaskHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(c, "task not found")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(c, "you do not have access to this task")
	default:
		response.InternalError(c)
	}
}
