package routes

import (
	"time"

	"task-management-api/internal/handler"
	"task-management-api/internal/middleware"
	"task-management-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	taskHandler *handler.TaskHandler,
	authService service.AuthService,
) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Auth routes (public)
	app.Post("/auth/register", authHandler.Register)
	app.Post("/auth/login", authHandler.Login)

	// Task routes (protected)
	api := app.Group("/tasks", middleware.AuthMiddleware(authService))
	api.Post("/", taskHandler.Create)
	api.Get("/", taskHandler.List)
	api.Get("/:id", taskHandler.GetByID)
	api.Put("/:id", taskHandler.Update)
	api.Delete("/:id", taskHandler.Delete)
}
