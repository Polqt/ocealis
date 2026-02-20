package handler

import (
	"github.com/Polqt/ocealis/api/middleware"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type createUserRequest struct {
	Nickname  string `json:"nickname" validate:"required,min=3,max=20"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url"`
}

type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// Create a new user
// Create an anonymous user with a nickname and optional avatar URL. The nickname must be between 3 and 20 characters, and the avatar URL must be a valid URL if provided.
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	var req createUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	user, err := h.svc.CreateUser(c.Context(), service.CreateUserInput{
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create user")
	}

	token, err := middleware.IssueToken(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not issue token")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":  user,
		"token": token,
	})
}
