package presentation_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/klnp-studio/backend/internal/modules/document/application"
	"github.com/klnp-studio/backend/internal/modules/document/domain"
	"github.com/klnp-studio/backend/internal/modules/document/presentation"
	"github.com/klnp-studio/backend/internal/security/middleware"
)

type stubRepo struct {
	rev int64
}

func (s *stubRepo) Get(context.Context, string, string, string) (*domain.PageDocument, error) {
	return &domain.PageDocument{PageID: "p", Revision: s.rev, Content: json.RawMessage(domain.EmptyDocument)}, nil
}

func (s *stubRepo) Mutate(_ context.Context, _, _, _ string, base int64, _ string, fn application.MutateFunc) (*domain.PageDocument, error) {
	if base != s.rev {
		return nil, &domain.ConflictError{CurrentRevision: s.rev}
	}
	content, _, err := fn(&domain.PageDocument{Revision: s.rev, Content: json.RawMessage(domain.EmptyDocument)})
	if err != nil {
		return nil, err
	}
	s.rev++
	return &domain.PageDocument{PageID: "p", Revision: s.rev, Content: content}, nil
}

func (s *stubRepo) ListOperations(context.Context, string, string, string, int64, int) ([]domain.Operation, error) {
	return nil, nil
}

func newApp(user string) *fiber.App {
	h := presentation.NewHandler(application.NewService(&stubRepo{rev: 3}))
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if user != "" {
			c.Locals(middleware.UserIDKey, user)
		}
		return c.Next()
	})
	g := app.Group("/workspaces/:workspaceId/projects/:projectId/pages/:pageId/document")
	g.Get("", h.Get)
	g.Patch("", h.ApplyPatch)
	g.Get("/operations", h.ListOperations)
	return app
}

const base = "/workspaces/w/projects/pr/pages/p/document"

func do(t *testing.T, app *fiber.App, method, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, base, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	return res.StatusCode, out
}

func TestGet(t *testing.T) {
	code, out := do(t, newApp("u"), "GET", "")
	assert.Equal(t, 200, code)
	assert.Equal(t, true, out["success"])
	assert.EqualValues(t, 3, out["data"].(map[string]any)["revision"])
}

func TestPatch_Success(t *testing.T) {
	body := `{"base_revision":3,"patch":[{"op":"add","path":"/nodes/a","value":{"id":"a","kind":"rect","children":[]}},{"op":"add","path":"/rootNodeIds/-","value":"a"}]}`
	code, out := do(t, newApp("u"), "PATCH", body)
	assert.Equal(t, 200, code)
	assert.EqualValues(t, 4, out["data"].(map[string]any)["revision"])
}

func TestPatch_ConflictReturns409WithCurrentRevision(t *testing.T) {
	body := `{"base_revision":1,"patch":[{"op":"add","path":"/x","value":1}]}`
	code, out := do(t, newApp("u"), "PATCH", body)
	assert.Equal(t, 409, code)
	assert.Equal(t, "REVISION_CONFLICT", out["error"].(map[string]any)["code"])
	assert.EqualValues(t, 3, out["meta"].(map[string]any)["current_revision"])
}

func TestPatch_InvalidGraphReturns422(t *testing.T) {
	body := `{"base_revision":3,"patch":[{"op":"add","path":"/rootNodeIds/-","value":"ghost"}]}`
	code, out := do(t, newApp("u"), "PATCH", body)
	assert.Equal(t, 422, code)
	assert.Equal(t, "INVALID_DOCUMENT", out["error"].(map[string]any)["code"])
}

func TestPatch_BadPatchReturns400(t *testing.T) {
	code, out := do(t, newApp("u"), "PATCH", `{"base_revision":3,"patch":[]}`)
	assert.Equal(t, 400, code)
	assert.Equal(t, "INVALID_PATCH", out["error"].(map[string]any)["code"])
}

func TestPatch_RequiresAuthenticatedUser(t *testing.T) {
	code, _ := do(t, newApp(""), "PATCH", `{"base_revision":3,"patch":[]}`)
	assert.Equal(t, 401, code)
}
