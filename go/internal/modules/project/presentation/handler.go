package presentation

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/klnp-studio/backend/internal/modules/project/application"
	"github.com/klnp-studio/backend/internal/modules/project/domain"
	"github.com/klnp-studio/backend/internal/security/middleware"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	userID := middleware.GetUserID(c)

	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("BAD_REQUEST", "invalid JSON payload", nil),
		)
	}

	project, err := h.service.Create(c.Context(), domain.CreateProjectInput{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			return c.Status(fiber.StatusBadRequest).JSON(
				response.Fail("VALIDATION_ERROR", err.Error(), nil),
			)
		case errors.Is(err, domain.ErrQuotaExceeded):
			return c.Status(fiber.StatusForbidden).JSON(
				response.Fail("PAGE_QUOTA_EXCEEDED", "workspace page quota exceeded for current plan", nil),
			)
		case errors.Is(err, domain.ErrSlugConflict):
			return c.Status(fiber.StatusConflict).JSON(
				response.Fail("SLUG_CONFLICT", "a project with this name already exists", nil),
			)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.Fail("INTERNAL_ERROR", "failed to create project", nil),
			)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(
		response.Success(ToProjectResponse(*project), nil),
	)
}

func (h *Handler) List(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	keyword := c.Query("keyword")

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	projects, total, err := h.service.List(c.Context(), workspaceID, keyword, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to fetch projects", nil),
		)
	}

	result := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		result[i] = ToProjectResponse(p)
	}

	return c.JSON(response.Success(result, PaginationMeta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}))
}

func (h *Handler) Get(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	p, err := h.service.Get(c.Context(), workspaceID, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "project not found", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to fetch project", nil),
		)
	}

	return c.JSON(response.Success(ToProjectResponse(*p), nil))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	var req UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("BAD_REQUEST", "invalid JSON payload", nil),
		)
	}

	p, err := h.service.Update(c.Context(), workspaceID, domain.UpdateProjectInput{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "project not found", nil),
			)
		case errors.Is(err, domain.ErrInvalidInput):
			return c.Status(fiber.StatusBadRequest).JSON(
				response.Fail("VALIDATION_ERROR", err.Error(), nil),
			)
		case errors.Is(err, domain.ErrSlugConflict):
			return c.Status(fiber.StatusConflict).JSON(
				response.Fail("SLUG_CONFLICT", "a project with this name already exists", nil),
			)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.Fail("INTERNAL_ERROR", "failed to update project", nil),
			)
		}
	}

	return c.JSON(response.Success(ToProjectResponse(*p), nil))
}

func (h *Handler) Archive(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	if err := h.service.Archive(c.Context(), workspaceID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "project not found", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to archive project", nil),
		)
	}

	return c.JSON(response.Success(fiber.Map{"archived": true}, nil))
}

func (h *Handler) Restore(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	if err := h.service.Restore(c.Context(), workspaceID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "project not found", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to restore project", nil),
		)
	}

	return c.JSON(response.Success(fiber.Map{"restored": true}, nil))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	if err := h.service.Delete(c.Context(), workspaceID, projectID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "project not found", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to delete project", nil),
		)
	}

	return c.JSON(response.Success(fiber.Map{"deleted": true}, nil))
}
