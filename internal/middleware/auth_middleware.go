package middleware

import (
	"strings"

	"task-management-api/internal/domain"
	"task-management-api/internal/service"
	"task-management-api/internal/util"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return util.SendError(c, fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return util.SendError(c, fiber.StatusUnauthorized, "invalid authorization header format")
		}

		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return util.SendError(c, fiber.StatusUnauthorized, "invalid or expired token")
		}

		// Store user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)

		return c.Next()
	}
}

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("userRole").(string)
		if !ok || role != string(domain.RoleAdmin) {
			return util.SendError(c, fiber.StatusForbidden, "admin access required")
		}
		return c.Next()
	}
}
