package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/internal/http/middleware"
	"github.com/iperez/new-expenses-go/internal/service"
	"github.com/iperez/new-expenses-go/pkg/response"
)

// CategoryHandler exposes the CRUD endpoints for categories.
type CategoryHandler struct {
	categories *service.CategoryService
}

func NewCategoryHandler(categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) Register(router fiber.Router) {
	router.Get("/", middleware.RequireAuth(), h.List)
	router.Post("/", middleware.RequireAuth(), h.Create)
	router.Get("/:categoryId", middleware.RequireAuth(), h.Get)
	router.Delete("/:categoryId", middleware.RequireAuth(), h.Delete)
}

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	categories, err := h.categories.List(c.UserContext(), middleware.UserID(c))
	if err != nil {
		return err
	}

	responses := make([]categoryResponse, 0, len(categories))
	for idx := range categories {
		responses = append(responses, newCategoryResponse(&categories[idx]))
	}

	return c.JSON(response.Success(responses))
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var payload service.CreateCategoryInput
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	category, err := h.categories.Create(c.UserContext(), middleware.UserID(c), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(newCategoryResponse(category)))
}

func (h *CategoryHandler) Get(c *fiber.Ctx) error {
	category, err := h.categories.GetByID(c.UserContext(), middleware.UserID(c), c.Params("categoryId"))
	if err != nil {
		return err
	}

	return c.JSON(response.Success(newCategoryResponse(category)))
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	deleteTransactions := c.QueryBool("deleteTransactions")

	if err := h.categories.Delete(c.UserContext(), middleware.UserID(c), c.Params("categoryId"), deleteTransactions); err != nil {
		return err
	}

	return c.JSON(response.Success(nil))
}

type categoryResponse struct {
	CategoryID string `json:"categoryId"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Note       string `json:"note"`
}

func newCategoryResponse(category *models.Category) categoryResponse {
	return categoryResponse{
		CategoryID: category.CategoryID,
		Type:       string(category.Type),
		Name:       category.Name,
		Note:       category.Note,
	}
}
