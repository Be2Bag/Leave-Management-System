package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github/be2bag/leave-management-system/internal/adapters/dto"
	"github/be2bag/leave-management-system/internal/core/domain"
)

var errorStatusMap = map[error]int{
	// 400 Bad Request — ข้อมูลที่ส่งมาไม่ถูกต้อง
	domain.ErrInvalidLeaveType: fiber.StatusBadRequest,
	domain.ErrInvalidDateRange: fiber.StatusBadRequest,

	// 401 Unauthorized — ยืนยันตัวตนไม่สำเร็จ
	domain.ErrInvalidCredentials: fiber.StatusUnauthorized,
	domain.ErrUnauthorized:       fiber.StatusUnauthorized,

	// 403 Forbidden — ไม่มีสิทธิ์ดำเนินการ
	domain.ErrSelfApproval: fiber.StatusForbidden,

	// 404 Not Found — ไม่พบข้อมูล
	domain.ErrUserNotFound:         fiber.StatusNotFound,
	domain.ErrRequestNotFound:      fiber.StatusNotFound,
	domain.ErrLeaveBalanceNotFound: fiber.StatusNotFound,

	// 409 Conflict — ข้อมูลขัดแย้ง
	domain.ErrOverlappingLeave:        fiber.StatusConflict,
	domain.ErrRequestNotPending:       fiber.StatusConflict,
	domain.ErrRequestAlreadyProcessed: fiber.StatusConflict,

	// 422 Unprocessable Entity — เงื่อนไขทาง business ไม่ผ่าน
	domain.ErrInsufficientBalance: fiber.StatusUnprocessableEntity,
}

// handleDomainError แปลง domain error เป็น HTTP response
func handleDomainError(c *fiber.Ctx, err error) error {
	for domainErr, status := range errorStatusMap {
		if errors.Is(err, domainErr) {
			return c.Status(status).JSON(dto.NewErrorResponse(domainErr.Error()))
		}
	}

	return c.Status(fiber.StatusInternalServerError).JSON(
		dto.NewErrorResponse("เกิดข้อผิดพลาดภายในระบบ"),
	)
}

// handleBodyParseError จัดการ error จากการ parse request body
func handleBodyParseError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(
		dto.NewErrorResponse("ข้อมูลที่ส่งมาไม่ถูกต้อง กรุณาตรวจสอบรูปแบบ JSON"),
	)
}

// handleValidationError จัดการ validation errors
func handleValidationError(c *fiber.Ctx, errs []string) error {
	return c.Status(fiber.StatusBadRequest).JSON(
		dto.NewValidationErrorResponse(errs),
	)
}
