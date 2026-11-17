package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	workspaceDomain "github.com/klnp-studio/backend/internal/modules/workspace/domain"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type Permission string

const (
	PermissionProjectRead   Permission = "project.read"
	PermissionProjectWrite  Permission = "project.write"
	PermissionProjectDelete Permission = "project.delete"

	PermissionPageRead   Permission = "page.read"
	PermissionPageWrite  Permission = "page.write"
	PermissionPageDelete Permission = "page.delete"

	PermissionPlanManage Permission = "plan.manage"
)

var rolePermissions = map[workspaceDomain.WorkspaceRole]map[Permission]bool{
	workspaceDomain.RoleOwner: {
		PermissionProjectRead:   true,
		PermissionProjectWrite:  true,
		PermissionProjectDelete: true,
		PermissionPageRead:      true,
		PermissionPageWrite:     true,
		PermissionPageDelete:    true,
		PermissionPlanManage:    true,
	},
	workspaceDomain.RoleAdmin: {
		PermissionProjectRead:   true,
		PermissionProjectWrite:  true,
		PermissionProjectDelete: true,
		PermissionPageRead:      true,
		PermissionPageWrite:     true,
		PermissionPageDelete:    true,
	},
	workspaceDomain.RoleEditor: {
		PermissionProjectRead: true,
		PermissionPageRead:    true,
		PermissionPageWrite:   true,
		PermissionPageDelete:  true,
	},
	workspaceDomain.RoleViewer: {
		PermissionProjectRead: true,
		PermissionPageRead:    true,
	},
}

func HasPermission(role workspaceDomain.WorkspaceRole, perm Permission) bool {
	if m, ok := rolePermissions[role]; ok {
		return m[perm]
	}
	return false
}

const (
	WorkspaceIDKey   = "workspace_id"
	WorkspaceRoleKey = "workspace_role"
)

func RequireWorkspacePermission(repo workspaceDomain.WorkspaceRepository, perm Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr := GetUserID(c)
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				response.Fail("UNAUTHENTICATED", "authentication required", nil),
			)
		}

		workspaceIDStr := c.Params("workspaceId")
		wsID, err := uuid.Parse(workspaceIDStr)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "workspace not found", nil),
			)
		}

		member, err := repo.GetMember(c.Context(), wsID, uid)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.Fail("NOT_FOUND", "workspace not found", nil),
			)
		}

		if !HasPermission(member.Role, perm) {
			return c.Status(fiber.StatusForbidden).JSON(
				response.Fail("FORBIDDEN", "insufficient workspace permissions", nil),
			)
		}

		c.Locals(WorkspaceIDKey, workspaceIDStr)
		c.Locals(WorkspaceRoleKey, string(member.Role))

		return c.Next()
	}
}

