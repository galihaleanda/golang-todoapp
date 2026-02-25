package handler

import (
	"errors"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/galihaleanda/todo-app/internal/middleware"
	"github.com/galihaleanda/todo-app/internal/service"
	"github.com/galihaleanda/todo-app/internal/validator"
	"github.com/galihaleanda/todo-app/pkg/response"
	"github.com/gin-gonic/gin"
)

// ProjectHandler exposes project CRUD endpoints.
type ProjectHandler struct {
	projectSvc *service.ProjectService
}

// NewProjectHandler creates a ProjectHandler.
func NewProjectHandler(projectSvc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectSvc: projectSvc}
}

// Create godoc
// @Summary Create a project
// @Tags projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body domain.CreateProjectRequest true "Project payload"
// @Success 201 {object} response.Envelope{data=domain.Project}
// @Router /projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	var req domain.CreateProjectRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	project, err := h.projectSvc.Create(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Created(c, project)
}

// List godoc
// @Summary List projects for current user
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Envelope{data=[]domain.Project}
// @Router /projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	projects, err := h.projectSvc.List(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, projects)
}

// GetByID godoc
// @Summary Get a project by ID
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path string true "Project UUID"
// @Success 200 {object} response.Envelope{data=domain.Project}
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project id", nil)
		return
	}

	project, err := h.projectSvc.GetByID(c.Request.Context(), id, middleware.CurrentUserID(c))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, project)
}

// Update godoc
// @Summary Update a project
// @Tags projects
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Project UUID"
// @Param body body domain.UpdateProjectRequest true "Update payload"
// @Success 200 {object} response.Envelope{data=domain.Project}
// @Router /projects/{id} [patch]
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project id", nil)
		return
	}

	var req domain.UpdateProjectRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	project, err := h.projectSvc.Update(c.Request.Context(), id, middleware.CurrentUserID(c), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, project)
}

// Delete godoc
// @Summary Delete a project
// @Tags projects
// @Security BearerAuth
// @Produce json
// @Param id path string true "Project UUID"
// @Success 200 {object} response.Envelope
// @Router /projects/{id} [delete]
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project id", nil)
		return
	}

	if err := h.projectSvc.Delete(c.Request.Context(), id, middleware.CurrentUserID(c)); err != nil {
		h.handleError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "project deleted"})
}

func (h *ProjectHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(c, "project not found")
	case errors.Is(err, domain.ErrForbidden):
		response.Forbidden(c, "you do not have access to this project")
	default:
		response.InternalError(c)
	}
}
