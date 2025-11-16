package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/iperez/new-expenses-go/internal/config"
	"github.com/iperez/new-expenses-go/internal/http/middleware"
	"github.com/iperez/new-expenses-go/internal/service"
	"github.com/iperez/new-expenses-go/pkg/apperror"
	"github.com/iperez/new-expenses-go/pkg/response"
)

// AuthHandler exposes login/logout endpoints.
type AuthHandler struct {
	auth *service.AuthService
	cfg  config.Config
}

func NewAuthHandler(auth *service.AuthService, cfg config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

func (h *AuthHandler) Register(router fiber.Router) {
	router.Post("/login", h.Login)
	router.Delete("/logout", middleware.RequireAuth(), h.Logout)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var payload loginRequest
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	payload.Email = strings.TrimSpace(payload.Email)
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Email == "" || payload.Password == "" {
		return apperror.New(apperror.ServerParamsMissing, "Email and password are required")
	}

	user, sessionID, err := h.auth.Login(c.UserContext(), payload.Email, payload.Password, middleware.SessionID(c))
	if err != nil {
		return err
	}

	h.setSessionCookie(c, sessionID)

	return c.JSON(response.Success(newUserResponse(user)))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	if err := h.auth.Logout(c.UserContext(), middleware.SessionID(c)); err != nil {
		return err
	}

	h.clearSessionCookie(c)

	return c.JSON(response.Success(nil))
}

func (h *AuthHandler) setSessionCookie(c *fiber.Ctx, sessionID string) {
	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    sessionID,
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   int(h.cfg.SessionTTL.Seconds()),
		Secure:   h.cfg.Env == "PROD",
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func (h *AuthHandler) clearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
	})
}
