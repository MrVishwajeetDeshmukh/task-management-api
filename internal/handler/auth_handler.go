package handler

import (
	"task-management-api/internal/domain"
	"task-management-api/internal/service"
	"task-management-api/internal/util"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := util.ValidateEmail(req.Email); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := util.ValidatePassword(req.Password); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := h.authService.Register(req)
	if err != nil {
		return util.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return util.SendSuccess(c, fiber.StatusCreated, user)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := util.ValidateEmail(req.Email); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := util.ValidatePassword(req.Password); err != nil {
		return util.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	response, err := h.authService.Login(req)
	if err != nil {
		return util.SendError(c, fiber.StatusUnauthorized, err.Error())
	}

	return util.SendSuccess(c, fiber.StatusOK, response)
}
