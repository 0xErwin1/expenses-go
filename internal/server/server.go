package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/iperez/new-expenses-go/internal/config"
	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/internal/http/handlers"
	"github.com/iperez/new-expenses-go/internal/http/middleware"
	"github.com/iperez/new-expenses-go/internal/service"
	"github.com/iperez/new-expenses-go/pkg/apperror"
	"github.com/iperez/new-expenses-go/pkg/response"
)

// Server wires the Fiber HTTP server with the services.
type Server struct {
	cfg config.Config
	app *fiber.App
}

// New bootstraps the HTTP server with every dependency wired.
func New(cfg config.Config, db *gorm.DB, redisClient *redis.Client) *Server {
	for _, model := range []interface{}{&models.User{}, &models.Category{}, &models.Transaction{}} {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
	}

	userService := service.NewUserService(db)
	categoryService := service.NewCategoryService(db)
	transactionService := service.NewTransactionService(db, categoryService)
	authService := service.NewAuthService(userService, redisClient, cfg.SessionTTL)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return errorHandler(c, err)
		},
	})

	app.Use(recover.New())
	app.Use(fiberLogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CorsOrigins, ","),
		AllowCredentials: true,
		AllowHeaders:     "Content-Type,Authorization",
	}))

	app.Use(middleware.Session(authService, cfg.SessionCookieName))

	api := app.Group("/api")

	handlers.NewHealthHandler().Register(api.Group("/health"))
	handlers.NewUserHandler(userService).Register(api.Group("/users"))
	handlers.NewAuthHandler(authService, cfg).Register(api.Group("/auth"))
	handlers.NewCategoryHandler(categoryService).Register(api.Group("/categories"))
	handlers.NewTransactionHandler(transactionService).Register(api.Group("/transactions"))

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.Error(apperror.ServerNotFound, nil))
	})

	return &Server{cfg: cfg, app: app}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	return s.app.Listen(fmt.Sprintf(":%d", s.cfg.Port))
}

// App exposes the underlying Fiber application (useful for integration tests).
func (s *Server) App() *fiber.App {
	return s.app
}

func errorHandler(c *fiber.Ctx, err error) error {
	// Log every error with method/path and body for easier debugging.
	log.Printf("error on %s %s: %v | body=%s", c.Method(), c.Path(), err, string(c.Body()))

	switch e := err.(type) {
	case apperror.AppError:
		def := apperror.Lookup(e.Code)
		return c.Status(def.HTTPStatus).JSON(response.Error(e.Code, e.Data))
	default:
		if appErr, ok := err.(*apperror.AppError); ok {
			def := apperror.Lookup(appErr.Code)
			return c.Status(def.HTTPStatus).JSON(response.Error(appErr.Code, appErr.Data))
		}
	}

	if fiberErr, ok := err.(*fiber.Error); ok {
		if fiberErr.Code == fiber.StatusNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.Error(apperror.ServerNotFound, nil))
		}

		return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
	}

	log.Printf("unexpected error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
}
