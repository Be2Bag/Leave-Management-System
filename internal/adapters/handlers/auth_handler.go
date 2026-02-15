package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github/be2bag/leave-management-system/internal/adapters/dto"
	"github/be2bag/leave-management-system/internal/core/ports"
	"github/be2bag/leave-management-system/pkg/validator"
)

type AuthHandler struct {
	authService ports.AuthService
	validate    *validator.Validator
}

func NewAuthHandler(authService ports.AuthService, validate *validator.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validate,
	}
}

// Login เข้าสู่ระบบ
//
//	@Summary		เข้าสู่ระบบ
//	@Description	ยืนยันตัวตนด้วยอีเมลและรหัสผ่าน จะได้รับ JWT token กลับมา
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.LoginRequest	true	"ข้อมูลสำหรับเข้าสู่ระบบ"
//	@Success		200	{object}	dto.APIResponse{data=dto.AuthResponse}
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return handleBodyParseError(c)
	}

	if errs := h.validate.Validate(req); errs != nil {
		return handleValidationError(c, errs)
	}

	token, user, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewSuccessResponse("เข้าสู่ระบบสำเร็จ", dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	}))
}
