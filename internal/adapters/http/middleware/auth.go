package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github/be2bag/leave-management-system/internal/adapters/dto"
	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
)

const bearerPrefix = "Bearer "

func AuthMiddleware(tokenService ports.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return unauthorizedResponse(c, "กรุณาระบุ Authorization header")
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return unauthorizedResponse(c, "รูปแบบ Authorization header ไม่ถูกต้อง (ต้องเป็น Bearer <token>)")
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			return unauthorizedResponse(c, "ไม่พบ token")
		}

		claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			return unauthorizedResponse(c, "token ไม่ถูกต้องหรือหมดอายุ")
		}

		c.Locals("userID", claims.UserID.String())
		c.Locals("email", claims.Email)
		c.Locals("role", string(claims.Role))

		return c.Next()
	}
}

func RoleMiddleware(allowedRoles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleStr, ok := c.Locals("role").(string)
		if !ok {
			return unauthorizedResponse(c, "ไม่พบข้อมูลบทบาทผู้ใช้")
		}

		userRole := domain.Role(roleStr)
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(
			dto.NewErrorResponse("ไม่มีสิทธิ์เข้าถึง endpoint นี้"),
		)
	}
}

func unauthorizedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(
		dto.NewErrorResponse(message),
	)
}
