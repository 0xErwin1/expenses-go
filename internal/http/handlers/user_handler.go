package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/internal/http/middleware"
	"github.com/iperez/new-expenses-go/internal/service"
	"github.com/iperez/new-expenses-go/pkg/apperror"
	"github.com/iperez/new-expenses-go/pkg/response"
)

// UserHandler wires HTTP requests with the user service.
type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Register(router fiber.Router) {
	router.Post("/", h.Create)
	router.Get("/", middleware.RequireAuth(), h.Me)
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var payload service.CreateUserInput
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	user, err := h.users.Create(c.UserContext(), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(newUserResponse(user)))
}

func (h *UserHandler) Me(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	if userID == "" {
		return apperror.New(apperror.AuthNeedLogin, nil)
	}

	user, err := h.users.GetByID(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return c.JSON(response.Success(newUserResponse(user)))
}

type userResponse struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func newUserResponse(user *models.User) userResponse {
	return userResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}
