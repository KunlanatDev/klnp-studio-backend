package presentation

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/klnp-studio/backend/internal/modules/page/application"
	"github.com/klnp-studio/backend/internal/modules/page/domain"
	"github.com/klnp-studio/backend/internal/security/middleware"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func mapPageError(err error) (int, response.Response) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return fiber.StatusNotFound, response.Fail("NOT_FOUND", "page not found", nil)
	case errors.Is(err, domain.ErrInvalidInput):
		return fiber.StatusBadRequest, response.Fail("INVALID_REQUEST", err.Error(), nil)
	case errors.Is(err, domain.ErrPathConflict):
		return fiber.StatusConflict, response.Fail("PATH_CONFLICT", "page slug or path already exists in this project", nil)
	case errors.Is(err, domain.ErrQuotaExceeded):
		return fiber.StatusConflict, response.Fail("QUOTA_EXCEEDED", "Your workspace has reached its page limit.", []response.ErrorDetail{
			{Field: "pages", Message: "Free plan limit reached (maximum 3 pages per workspace)"},
		})
	case errors.Is(err, domain.ErrCannotDeleteHome):
		return fiber.StatusBadRequest, response.Fail("CANNOT_DELETE_HOME", "cannot delete the active home page of a project", nil)
	default:
		return fiber.StatusInternalServerError, response.Fail("INTERNAL_ERROR", "an unexpected error occurred", nil)
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	actorID := middleware.GetUserID(c)

	if actorID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "authentication required", nil),
		)
	}

	var req CreatePageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("INVALID_REQUEST", "invalid JSON payload", nil),
		)
	}

	page, err := h.service.Create(c.Context(), domain.CreatePageInput{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Name:        req.Name,
		Slug:        req.Slug,
		CreatedBy:   actorID,
	})
	if err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.Status(fiber.StatusCreated).JSON(
		response.Success(ToPageResponse(*page), nil),
	)
}

func (h *Handler) List(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	pages, err := h.service.List(c.Context(), workspaceID, projectID)
	if err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	result := make([]PageResponse, len(pages))
	for i, p := range pages {
		result[i] = ToPageResponse(p)
	}

	return c.JSON(response.Success(result, nil))
}

func (h *Handler) Get(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")

	p, err := h.service.Get(c.Context(), workspaceID, projectID, pageID)
	if err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.JSON(response.Success(ToPageResponse(*p), nil))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")

	var req UpdatePageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("INVALID_REQUEST", "invalid JSON payload", nil),
		)
	}

	p, err := h.service.Update(c.Context(), domain.UpdatePageInput{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		PageID:      pageID,
		Name:        req.Name,
		Slug:        req.Slug,
	})
	if err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.JSON(response.Success(ToPageResponse(*p), nil))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")

	if err := h.service.Delete(c.Context(), workspaceID, projectID, pageID); err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.JSON(response.Success(fiber.Map{"deleted": true}, nil))
}

func (h *Handler) Duplicate(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")
	actorID := middleware.GetUserID(c)

	if actorID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "authentication required", nil),
		)
	}

	p, err := h.service.Duplicate(c.Context(), workspaceID, projectID, pageID, actorID)
	if err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.Status(fiber.StatusCreated).JSON(
		response.Success(ToPageResponse(*p), nil),
	)
}

func (h *Handler) SetHome(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")

	if err := h.service.SetHome(c.Context(), workspaceID, projectID, pageID); err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.JSON(response.Success(fiber.Map{"is_home": true}, nil))
}

func (h *Handler) Reorder(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")

	var req ReorderPagesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("INVALID_REQUEST", "invalid JSON payload", nil),
		)
	}

	if err := h.service.Reorder(c.Context(), domain.ReorderPagesInput{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		PageIDs:     req.PageIDs,
	}); err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	// Refetch ordered list
	pages, err := h.service.List(c.Context(), workspaceID, projectID)
	if err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	result := make([]PageResponse, len(pages))
	for i, p := range pages {
		result[i] = ToPageResponse(p)
	}

	return c.JSON(response.Success(result, nil))
}

func (h *Handler) Archive(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")

	if err := h.service.Archive(c.Context(), workspaceID, projectID, pageID); err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.JSON(response.Success(fiber.Map{"archived": true}, nil))
}

func (h *Handler) Restore(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	projectID := c.Params("projectId")
	pageID := c.Params("pageId")

	if err := h.service.Restore(c.Context(), workspaceID, projectID, pageID); err != nil {
		status, res := mapPageError(err)
		return c.Status(status).JSON(res)
	}

	return c.JSON(response.Success(fiber.Map{"restored": true}, nil))
}
