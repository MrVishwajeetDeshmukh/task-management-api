package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-management-api/internal/config"
	"task-management-api/internal/handler"
	"task-management-api/internal/repository"
	"task-management-api/internal/routes"
	"task-management-api/internal/service"
	"task-management-api/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	taskRepo := repository.NewTaskRepository(db.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)
	taskService := service.NewTaskService(taskRepo)
	workerService := service.NewWorkerService(taskRepo, cfg)

	// Start worker service with context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workerService.Start(ctx)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService, workerService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Setup Routes
	routes.SetupRoutes(app, authHandler, taskHandler, authService)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		cancel() // Cancel worker context

		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"message": err.Error(),
	})
}
