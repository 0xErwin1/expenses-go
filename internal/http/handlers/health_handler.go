package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/iperez/new-expenses-go/pkg/response"
)

// HealthHandler exposes a ready check endpoint.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Register(router fiber.Router) {
	router.Get("/", h.Health)
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(response.Success("Up & running ;)!"))
}
