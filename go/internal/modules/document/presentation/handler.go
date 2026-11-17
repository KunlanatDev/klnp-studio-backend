package presentation

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/klnp-studio/backend/internal/modules/document/application"
	"github.com/klnp-studio/backend/internal/modules/document/domain"
	"github.com/klnp-studio/backend/internal/security/middleware"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func pageRef(c *fiber.Ctx) application.PageRef {
	return application.PageRef{
		WorkspaceID: c.Params("workspaceId"),
		ProjectID:   c.Params("projectId"),
		PageID:      c.Params("pageId"),
	}
}

func mapError(err error) (int, response.Response) {
	var conflict *domain.ConflictError
	var validation *domain.ValidationError

	switch {
	case errors.As(err, &conflict):
		res := response.Fail("REVISION_CONFLICT", "document was changed by someone else; reload and retry", nil)
		res.Meta = fiber.Map{"current_revision": conflict.CurrentRevision}
		return fiber.StatusConflict, res
	case errors.As(err, &validation):
		details := make([]response.ErrorDetail, 0, len(validation.Issues))
		for i, is := range validation.Issues {
			if i == 20 {
				break
			}
			details = append(details, response.ErrorDetail{Field: is.NodeID, Message: is.Code + ": " + is.Message})
		}
		return fiber.StatusUnprocessableEntity, response.Fail("INVALID_DOCUMENT", "document violates design object graph rules", details)
	case errors.Is(err, domain.ErrNotFound):
		return fiber.StatusNotFound, response.Fail("NOT_FOUND", "page not found", nil)
	case errors.Is(err, domain.ErrInvalidPatch):
		return fiber.StatusBadRequest, response.Fail("INVALID_PATCH", err.Error(), nil)
	case errors.Is(err, domain.ErrDocumentTooLarge):
		return fiber.StatusRequestEntityTooLarge, response.Fail("DOCUMENT_TOO_LARGE", "document exceeds size limit", nil)
	case errors.Is(err, domain.ErrInvalidInput):
		return fiber.StatusBadRequest, response.Fail("INVALID_REQUEST", err.Error(), nil)
	default:
		return fiber.StatusInternalServerError, response.Fail("INTERNAL_ERROR", "an unexpected error occurred", nil)
	}
}

// Get returns the current document and revision.
func (h *Handler) Get(c *fiber.Ctx) error {
	doc, err := h.service.Get(c.Context(), pageRef(c))
	if err != nil {
		status, res := mapError(err)
		return c.Status(status).JSON(res)
	}
	return c.JSON(response.Success(ToDocumentResponse(*doc), nil))
}

// ApplyPatch applies a JSON Patch against base_revision.
func (h *Handler) ApplyPatch(c *fiber.Ctx) error {
	actorID := middleware.GetUserID(c)
	if actorID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "authentication required", nil))
	}

	var req ApplyPatchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("INVALID_REQUEST", "invalid JSON payload", nil))
	}

	doc, err := h.service.ApplyPatch(c.Context(), application.ApplyPatchInput{
		Page: pageRef(c), ActorID: actorID, BaseRevision: req.BaseRevision, Patch: req.Patch,
	})
	if err != nil {
		status, res := mapError(err)
		return c.Status(status).JSON(res)
	}
	return c.JSON(response.Success(ToPatchResultResponse(*doc), nil))
}

// ListOperations returns operations with revision > ?after (ascending).
func (h *Handler) ListOperations(c *fiber.Ctx) error {
	after := int64(c.QueryInt("after", 0))
	limit := c.QueryInt("limit", application.DefaultOperationsLimit)

	ops, err := h.service.ListOperations(c.Context(), pageRef(c), after, limit)
	if err != nil {
		status, res := mapError(err)
		return c.Status(status).JSON(res)
	}

	out := make([]OperationResponse, len(ops))
	for i, o := range ops {
		out[i] = ToOperationResponse(o)
	}
	return c.JSON(response.Success(out, nil))
}
