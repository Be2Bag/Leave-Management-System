package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github/be2bag/leave-management-system/internal/adapters/dto"
	"github/be2bag/leave-management-system/internal/adapters/handlers"
	"github/be2bag/leave-management-system/internal/adapters/http/middleware"
	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
)

func SetupRouter(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	leaveHandler *handlers.LeaveHandler,
	tokenService ports.TokenService,
) {
	app.Use(middleware.SecurityHeaders())

	app.Get("/health", healthCheck)

	api := app.Group("/api/v1")

	setupAuthRoutes(api, authHandler)

	protected := api.Group("", middleware.AuthMiddleware(tokenService))
	setupLeaveRoutes(protected, leaveHandler)
	setupManagerRoutes(protected, leaveHandler)
}

const authRateLimitMax = 10

func setupAuthRoutes(router fiber.Router, h *handlers.AuthHandler) {
	authLimiter := limiter.New(limiter.Config{
		Max:        authRateLimitMax,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // จำกัดต่อ IP address
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(
				dto.NewErrorResponse("คำขอมากเกินไป กรุณาลองอีกครั้งในภายหลัง"),
			)
		},
	})

	auth := router.Group("/auth", authLimiter)
	auth.Post("/login", h.Login) // เข้าสู่ระบบ
}

func setupLeaveRoutes(router fiber.Router, h *handlers.LeaveHandler) {
	leaves := router.Group("/leaves")
	leaves.Post("/", h.Submit)                  // ยื่นใบลา
	leaves.Get("/my-requests", h.GetMyRequests) // ดูประวัติใบลา
	leaves.Get("/my-balance", h.GetMyBalance)   // ดูยอดวันลาคงเหลือ
}

func setupManagerRoutes(router fiber.Router, h *handlers.LeaveHandler) {
	manager := router.Group("/manager", middleware.RoleMiddleware(domain.RoleManager))
	manager.Get("/pending-requests", h.GetPendingRequests) // ดูใบลารอการอนุมัติ
	manager.Post("/requests/:id/approve", h.Approve)       // อนุมัติใบลา
	manager.Post("/requests/:id/reject", h.Reject)         // ปฏิเสธใบลา
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "Leave Management System API",
	})
}
