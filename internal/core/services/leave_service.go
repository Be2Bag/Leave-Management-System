package services

import (
	"context"
	"fmt"
	"time"

	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
)

type leaveService struct {
	requestRepo ports.LeaveRequestRepository
	balanceRepo ports.LeaveBalanceRepository
}

func NewLeaveService(
	requestRepo ports.LeaveRequestRepository,
	balanceRepo ports.LeaveBalanceRepository,
) ports.LeaveService {
	return &leaveService{
		requestRepo: requestRepo,
		balanceRepo: balanceRepo,
	}
}

// Submit ยื่นใบลาใหม่ — ตรวจสอบเงื่อนไขทั้งหมดก่อนสร้าง
func (s *leaveService) Submit(
	ctx context.Context,
	userID domain.ID,
	leaveType domain.LeaveType,
	startDate, endDate time.Time,
	reason string,
) (*domain.LeaveRequest, error) {
	if !leaveType.IsValid() {
		return nil, domain.ErrInvalidLeaveType
	}

	if endDate.Before(startDate) {
		return nil, domain.ErrInvalidDateRange
	}

	request := domain.NewLeaveRequest(userID, leaveType, startDate, endDate, reason)

	if err := s.checkOverlap(ctx, userID, startDate, endDate); err != nil {
		return nil, err
	}

	if err := s.balanceRepo.ReservePending(ctx, userID, leaveType, startDate.Year(), request.TotalDays); err != nil {
		return nil, err
	}

	if err := s.requestRepo.Create(ctx, request); err != nil {
		// Rollback: ปล่อยวันลาที่จองไว้กลับ
		if rbErr := s.balanceRepo.ReleasePending(ctx, userID, leaveType, startDate.Year(), request.TotalDays); rbErr != nil {
			return nil, fmt.Errorf("บันทึกใบลาล้มเหลว: %w (rollback ล้มเหลว: %v)", err, rbErr)
		}
		return nil, fmt.Errorf("บันทึกใบลาล้มเหลว: %w", err)
	}

	return request, nil
}

// checkOverlap ตรวจสอบว่าวันลาซ้ำซ้อนกับใบลาอื่นหรือไม่
func (s *leaveService) checkOverlap(ctx context.Context, userID domain.ID, startDate, endDate time.Time) error {
	hasOverlap, err := s.requestRepo.HasOverlap(ctx, userID, startDate, endDate, nil)
	if err != nil {
		return fmt.Errorf("ตรวจสอบวันลาซ้ำซ้อนล้มเหลว: %w", err)
	}
	if hasOverlap {
		return domain.ErrOverlappingLeave
	}
	return nil
}

// GetMyRequests ดูประวัติใบลาทั้งหมดของพนักงาน (รองรับ pagination)
func (s *leaveService) GetMyRequests(
	ctx context.Context,
	userID domain.ID,
	params domain.PaginationParams,
) (*domain.PaginatedResult[domain.LeaveRequest], error) {
	result, err := s.requestRepo.FindByUserID(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("ดึงข้อมูลใบลาล้มเหลว: %w", err)
	}
	return result, nil
}

// GetMyBalance ดูยอดวันลาคงเหลือทั้งหมดของพนักงาน
func (s *leaveService) GetMyBalance(ctx context.Context, userID domain.ID) ([]domain.LeaveBalance, error) {
	balances, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ดึงข้อมูลยอดวันลาล้มเหลว: %w", err)
	}
	return balances, nil
}

// GetPendingRequests ดูใบลาที่รอการอนุมัติ (สำหรับผู้จัดการ, รองรับ pagination)
func (s *leaveService) GetPendingRequests(
	ctx context.Context,
	params domain.PaginationParams,
) (*domain.PaginatedResult[domain.LeaveRequest], error) {
	result, err := s.requestRepo.FindByStatus(ctx, domain.LeaveStatusPending, params)
	if err != nil {
		return nil, fmt.Errorf("ดึงข้อมูลใบลารอการอนุมัติล้มเหลว: %w", err)
	}
	return result, nil
}

// Approve อนุมัติใบลา — ย้ายวันลาจาก pending ไป used แบบ atomic
func (s *leaveService) Approve(ctx context.Context, requestID, reviewerID domain.ID, note string) error {
	request, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return err
	}

	if request.UserID == reviewerID {
		return domain.ErrSelfApproval
	}

	if err := request.Approve(reviewerID, note); err != nil {
		return err
	}

	if err := s.requestRepo.UpdateWithStatusCheck(ctx, request, domain.LeaveStatusPending); err != nil {
		return err
	}

	// ย้าย pending_days → used_days
	if err := s.balanceRepo.ConfirmPending(
		ctx, request.UserID, request.LeaveType, request.StartDate.Year(), request.TotalDays,
	); err != nil {
		// Rollback: คืนสถานะใบลากลับเป็น pending
		if rbErr := s.rollbackRequestStatus(ctx, request); rbErr != nil {
			return fmt.Errorf("ยืนยันยอดวันลาล้มเหลว: %w (rollback ล้มเหลว: %v)", err, rbErr)
		}
		return err
	}

	return nil
}

// Reject ปฏิเสธใบลา — ปล่อยวันลาที่จองไว้กลับคืน แบบ atomic
func (s *leaveService) Reject(ctx context.Context, requestID, reviewerID domain.ID, note string) error {
	request, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return err
	}
	if request.UserID == reviewerID {
		return domain.ErrSelfApproval
	}

	if err := request.Reject(reviewerID, note); err != nil {
		return err
	}

	if err := s.requestRepo.UpdateWithStatusCheck(ctx, request, domain.LeaveStatusPending); err != nil {
		return err
	}

	// ปล่อย pending_days กลับคืน
	if err := s.balanceRepo.ReleasePending(
		ctx, request.UserID, request.LeaveType, request.StartDate.Year(), request.TotalDays,
	); err != nil {
		// Rollback: คืนสถานะใบลากลับเป็น pending
		if rbErr := s.rollbackRequestStatus(ctx, request); rbErr != nil {
			return fmt.Errorf("ปล่อยวันลาที่จองไว้ล้มเหลว: %w (rollback ล้มเหลว: %v)", err, rbErr)
		}
		return err
	}

	return nil
}

// rollbackRequestStatus คืนสถานะใบลากลับเป็น pending เมื่อ balance update ล้มเหลว
func (s *leaveService) rollbackRequestStatus(ctx context.Context, request *domain.LeaveRequest) error {
	request.Status = domain.LeaveStatusPending
	request.ReviewerID = nil
	request.ReviewNote = ""
	request.ReviewedAt = nil
	return s.requestRepo.Update(ctx, request)
}
