package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github/be2bag/leave-management-system/internal/adapters/dto"
	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
	"github/be2bag/leave-management-system/pkg/validator"
)

const dateFormat = "2006-01-02"

type LeaveHandler struct {
	leaveService ports.LeaveService
	validate     *validator.Validator
}

func NewLeaveHandler(leaveService ports.LeaveService, validate *validator.Validator) *LeaveHandler {
	return &LeaveHandler{
		leaveService: leaveService,
		validate:     validate,
	}
}

// Submit ยื่นใบลาใหม่
//
//	@Summary		ยื่นใบลาใหม่
//	@Description	สร้างคำขอลาใหม่ ระบบจะตรวจสอบวันลาซ้ำซ้อนและยอดวันลาคงเหลือโดยอัตโนมัติ
//	@Tags			Leave
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.SubmitLeaveRequest	true	"ข้อมูลสำหรับยื่นใบลา"
//	@Success		201	{object}	dto.APIResponse{data=dto.LeaveRequestResponse}
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Failure		422	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/leaves [post]
func (h *LeaveHandler) Submit(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return handleDomainError(c, err)
	}

	var req dto.SubmitLeaveRequest
	if err = c.BodyParser(&req); err != nil {
		return handleBodyParseError(c)
	}

	if errs := h.validate.Validate(req); errs != nil {
		return handleValidationError(c, errs)
	}

	// แปลงวันที่จาก string เป็น time.Time
	startDate, endDate, err := parseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("รูปแบบวันที่ไม่ถูกต้อง กรุณาใช้ YYYY-MM-DD"),
		)
	}

	request, err := h.leaveService.Submit(
		c.Context(), userID, domain.LeaveType(req.LeaveType),
		startDate, endDate, req.Reason,
	)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(
		dto.NewSuccessResponse("ยื่นใบลาสำเร็จ", dto.ToLeaveRequestResponse(request)),
	)
}

// GetMyRequests ดูประวัติใบลาทั้งหมดของตนเอง
//
//	@Summary		ดูประวัติใบลา
//	@Description	ดึงข้อมูลใบลาทั้งหมดของผู้ใช้ที่เข้าสู่ระบบ (รองรับ pagination)
//	@Tags			Leave
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query	int	false	"หน้าที่ต้องการ (เริ่มจาก 1)"		default(1)
//	@Param			page_size	query	int	false	"จำนวนรายการต่อหน้า (สูงสุด 100)"	default(10)
//	@Success		200	{object}	dto.PaginatedAPIResponse{data=[]dto.LeaveRequestResponse}
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/leaves/my-requests [get]
func (h *LeaveHandler) GetMyRequests(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return handleDomainError(c, err)
	}

	params := parsePaginationParams(c)

	result, err := h.leaveService.GetMyRequests(c.Context(), userID, params)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(
		dto.NewPaginatedResponse(
			"ดึงข้อมูลใบลาสำเร็จ",
			dto.ToLeaveRequestResponses(result.Items),
			result.Page, result.PageSize, result.Total, result.TotalPages,
		),
	)
}

// GetMyBalance ดูยอดวันลาคงเหลือของตนเอง
//
//	@Summary		ดูยอดวันลาคงเหลือ
//	@Description	ดึงข้อมูลยอดวันลาคงเหลือทุกประเภทของผู้ใช้ที่เข้าสู่ระบบ
//	@Tags			Leave
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.APIResponse{data=[]dto.LeaveBalanceResponse}
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/leaves/my-balance [get]
func (h *LeaveHandler) GetMyBalance(c *fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return handleDomainError(c, err)
	}

	balances, err := h.leaveService.GetMyBalance(c.Context(), userID)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(
		dto.NewSuccessResponse("ดึงข้อมูลยอดวันลาสำเร็จ", dto.ToLeaveBalanceResponses(balances)),
	)
}

// GetPendingRequests ดูใบลาที่รอการอนุมัติ (เฉพาะ Manager)
//
//	@Summary		ดูใบลารอการอนุมัติ
//	@Description	ดึงข้อมูลใบลาทั้งหมดที่มีสถานะ pending สำหรับผู้จัดการ (รองรับ pagination)
//	@Tags			Manager
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query	int	false	"หน้าที่ต้องการ (เริ่มจาก 1)"		default(1)
//	@Param			page_size	query	int	false	"จำนวนรายการต่อหน้า (สูงสุด 100)"	default(10)
//	@Success		200	{object}	dto.PaginatedAPIResponse{data=[]dto.LeaveRequestResponse}
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/manager/pending-requests [get]
func (h *LeaveHandler) GetPendingRequests(c *fiber.Ctx) error {
	params := parsePaginationParams(c)

	result, err := h.leaveService.GetPendingRequests(c.Context(), params)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(
		dto.NewPaginatedResponse(
			"ดึงข้อมูลใบลารอการอนุมัติสำเร็จ",
			dto.ToLeaveRequestResponses(result.Items),
			result.Page, result.PageSize, result.Total, result.TotalPages,
		),
	)
}

// Approve อนุมัติใบลา (เฉพาะ Manager)
//
//	@Summary		อนุมัติใบลา
//	@Description	ผู้จัดการอนุมัติคำขอลา ระบบจะหักยอดวันลาของพนักงานโดยอัตโนมัติ
//	@Tags			Manager
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string					true	"รหัสใบลา (UUID)"
//	@Param			request	body	dto.ReviewLeaveRequest	true	"หมายเหตุจากผู้อนุมัติ"
//	@Success		200	{object}	dto.APIResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/manager/requests/{id}/approve [post]
func (h *LeaveHandler) Approve(c *fiber.Ctx) error {
	return h.reviewRequest(c, true)
}

// Reject ปฏิเสธใบลา (เฉพาะ Manager)
//
//	@Summary		ปฏิเสธใบลา
//	@Description	ผู้จัดการปฏิเสธคำขอลา ยอดวันลาของพนักงานจะไม่เปลี่ยนแปลง
//	@Tags			Manager
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string					true	"รหัสใบลา (UUID)"
//	@Param			request	body	dto.ReviewLeaveRequest	true	"หมายเหตุจากผู้อนุมัติ"
//	@Success		200	{object}	dto.APIResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Failure		500	{object}	dto.ErrorResponse
//	@Router			/api/v1/manager/requests/{id}/reject [post]
func (h *LeaveHandler) Reject(c *fiber.Ctx) error {
	return h.reviewRequest(c, false)
}

func (h *LeaveHandler) reviewRequest(c *fiber.Ctx, approve bool) error {
	reviewerID, err := getUserIDFromContext(c)
	if err != nil {
		return handleDomainError(c, err)
	}

	requestID, err := domain.ParseID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("รหัสใบลาไม่ถูกต้อง"),
		)
	}

	var req dto.ReviewLeaveRequest
	if err = c.BodyParser(&req); err != nil {
		return handleBodyParseError(c)
	}

	if approve {
		err = h.leaveService.Approve(c.Context(), requestID, reviewerID, req.Note)
	} else {
		err = h.leaveService.Reject(c.Context(), requestID, reviewerID, req.Note)
	}
	if err != nil {
		return handleDomainError(c, err)
	}

	message := "อนุมัติใบลาสำเร็จ"
	if !approve {
		message = "ปฏิเสธใบลาสำเร็จ"
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewSuccessResponse(message, nil))
}

func getUserIDFromContext(c *fiber.Ctx) (domain.ID, error) {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return domain.ID{}, domain.ErrUnauthorized
	}

	userID, err := domain.ParseID(userIDStr)
	if err != nil {
		return domain.ID{}, domain.ErrUnauthorized
	}

	return userID, nil
}

func parseDateRange(startStr, endStr string) (startDate, endDate time.Time, err error) {
	startDate, err = time.Parse(dateFormat, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err = time.Parse(dateFormat, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

func parsePaginationParams(c *fiber.Ctx) domain.PaginationParams {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		page = domain.DefaultPage
	}
	pageSize, err := strconv.Atoi(c.Query("page_size", "10"))
	if err != nil {
		pageSize = domain.DefaultPageSize
	}
	return domain.NewPaginationParams(page, pageSize)
}
