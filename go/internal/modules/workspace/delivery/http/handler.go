package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/workspace/application"
	"github.com/klnp-studio/backend/internal/security/middleware"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type WorkspaceHandler struct {
	useCase *application.WorkspaceUseCase
}

func NewWorkspaceHandler(useCase *application.WorkspaceUseCase) *WorkspaceHandler {
	return &WorkspaceHandler{useCase: useCase}
}

type CreateWorkspaceDTO struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateWorkspaceDTO struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Create(c *fiber.Ctx) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "invalid user id in context", nil),
		)
	}

	var dto CreateWorkspaceDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("VALIDATION_ERROR", "invalid request body", nil),
		)
	}

	res, err := h.useCase.CreateWorkspace(c.Context(), uid, application.CreateWorkspaceInput{
		Name: dto.Name,
		Slug: dto.Slug,
	})
	if err != nil {
		if errors.Is(err, appErrors.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(
				response.Fail("CONFLICT", "workspace slug already exists", nil),
			)
		}
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("VALIDATION_ERROR", err.Error(), nil),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(res, nil))
}

func (h *WorkspaceHandler) List(c *fiber.Ctx) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "invalid user id in context", nil),
		)
	}

	workspaces, err := h.useCase.ListWorkspaces(c.Context(), uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to list workspaces", nil),
		)
	}

	return c.JSON(response.Success(workspaces, nil))
}

func (h *WorkspaceHandler) Get(c *fiber.Ctx) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "invalid user id in context", nil),
		)
	}

	wsIDStr := c.Params("workspaceId")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.Fail("NOT_FOUND", "workspace not found", nil),
		)
	}

	ws, err := h.useCase.GetWorkspace(c.Context(), uid, wsID)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "workspace not found", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to fetch workspace", nil),
		)
	}

	return c.JSON(response.Success(ws, nil))
}

func (h *WorkspaceHandler) Update(c *fiber.Ctx) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "invalid user id in context", nil),
		)
	}

	wsIDStr := c.Params("workspaceId")
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.Fail("NOT_FOUND", "workspace not found", nil),
		)
	}

	var dto UpdateWorkspaceDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("VALIDATION_ERROR", "invalid request body", nil),
		)
	}

	ws, err := h.useCase.UpdateWorkspace(c.Context(), uid, wsID, dto.Name)
	if err != nil {
		if errors.Is(err, appErrors.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(
				response.Fail("FORBIDDEN", "owner permission required to update workspace", nil),
			)
		}
		if errors.Is(err, appErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "workspace not found", nil),
			)
		}
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("VALIDATION_ERROR", err.Error(), nil),
		)
	}

	return c.JSON(response.Success(ws, nil))
}
