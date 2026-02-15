package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github/be2bag/leave-management-system/internal/core/domain"
)

func TestLeaveService_Submit_Success(t *testing.T) {
	userID := domain.NewID()

	requestRepo := &mockLeaveRequestRepository{
		hasOverlapFn: func(_ context.Context, _ domain.ID, _, _ time.Time, _ *domain.ID) (bool, error) {
			return false, nil // ไม่มีวันลาซ้ำซ้อน
		},
	}
	balanceRepo := &mockLeaveBalanceRepository{
		reservePendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, _ float64) error {
			return nil // จองวันลาสำเร็จ
		},
	}

	svc := NewLeaveService(requestRepo, balanceRepo)

	startDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)

	request, err := svc.Submit(context.Background(), userID, domain.LeaveTypeSick, startDate, endDate, "ไม่สบาย")

	require.NoError(t, err)
	assert.Equal(t, userID, request.UserID)
	assert.Equal(t, domain.LeaveTypeSick, request.LeaveType)
	assert.Equal(t, domain.LeaveStatusPending, request.Status)
	assert.Equal(t, "ไม่สบาย", request.Reason)
	assert.Equal(t, float64(3), request.TotalDays)
}

func TestLeaveService_Submit_InvalidLeaveType(t *testing.T) {
	svc := NewLeaveService(&mockLeaveRequestRepository{}, &mockLeaveBalanceRepository{})

	startDate := time.Now()
	endDate := startDate.Add(24 * time.Hour)

	_, err := svc.Submit(context.Background(), domain.NewID(), domain.LeaveType("invalid"), startDate, endDate, "test")

	assert.ErrorIs(t, err, domain.ErrInvalidLeaveType)
}

func TestLeaveService_Submit_InvalidDateRange(t *testing.T) {
	svc := NewLeaveService(&mockLeaveRequestRepository{}, &mockLeaveBalanceRepository{})

	startDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC) // วันสิ้นสุดก่อนวันเริ่มต้น

	_, err := svc.Submit(context.Background(), domain.NewID(), domain.LeaveTypeSick, startDate, endDate, "test")

	assert.ErrorIs(t, err, domain.ErrInvalidDateRange)
}

func TestLeaveService_Submit_OverlappingLeave(t *testing.T) {
	requestRepo := &mockLeaveRequestRepository{
		hasOverlapFn: func(_ context.Context, _ domain.ID, _, _ time.Time, _ *domain.ID) (bool, error) {
			return true, nil // มีวันลาซ้ำซ้อน
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	startDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)

	_, err := svc.Submit(context.Background(), domain.NewID(), domain.LeaveTypeSick, startDate, endDate, "test")

	assert.ErrorIs(t, err, domain.ErrOverlappingLeave)
}

func TestLeaveService_Submit_InsufficientBalance(t *testing.T) {
	requestRepo := &mockLeaveRequestRepository{
		hasOverlapFn: func(_ context.Context, _ domain.ID, _, _ time.Time, _ *domain.ID) (bool, error) {
			return false, nil
		},
	}
	balanceRepo := &mockLeaveBalanceRepository{
		reservePendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, _ float64) error {
			// Atomic reservation ล้มเหลว — วันลาไม่เพียงพอ
			return domain.ErrInsufficientBalance
		},
	}

	svc := NewLeaveService(requestRepo, balanceRepo)

	startDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC) // 3 วัน

	_, err := svc.Submit(context.Background(), domain.NewID(), domain.LeaveTypeSick, startDate, endDate, "test")

	assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
}

func TestLeaveService_Approve_Success(t *testing.T) {
	employeeID := domain.NewID()
	managerID := domain.NewID()

	// สร้างใบลาตัวอย่าง
	request := domain.NewLeaveRequest(
		employeeID, domain.LeaveTypeAnnual,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC),
		"พักผ่อน",
	)

	var updatedRequest *domain.LeaveRequest

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
		updateWithStatusCheckFn: func(_ context.Context, r *domain.LeaveRequest, expectedStatus domain.LeaveStatus) error {
			if expectedStatus != domain.LeaveStatusPending {
				t.Errorf("expected status check for pending, got %s", expectedStatus)
			}
			updatedRequest = r
			return nil
		},
	}
	balanceRepo := &mockLeaveBalanceRepository{
		confirmPendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, _ float64) error {
			return nil // ย้ายจาก pending ไป used สำเร็จ
		},
	}

	svc := NewLeaveService(requestRepo, balanceRepo)

	err := svc.Approve(context.Background(), request.ID, managerID, "อนุมัติ")

	require.NoError(t, err)
	assert.Equal(t, domain.LeaveStatusApproved, updatedRequest.Status)
	assert.Equal(t, "อนุมัติ", updatedRequest.ReviewNote)
}

func TestLeaveService_Approve_AlreadyProcessed(t *testing.T) {
	// จำลองสถานการณ์: 2 manager approve พร้อมกัน — คนแรกสำเร็จ คนที่สองต้องได้ error
	employeeID := domain.NewID()
	managerID := domain.NewID()

	request := domain.NewLeaveRequest(
		employeeID, domain.LeaveTypeAnnual,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC),
		"พักผ่อน",
	)

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
		updateWithStatusCheckFn: func(_ context.Context, _ *domain.LeaveRequest, _ domain.LeaveStatus) error {
			return domain.ErrRequestAlreadyProcessed // จำลองว่ามีคนอื่น approve ไปก่อนแล้ว
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	err := svc.Approve(context.Background(), request.ID, managerID, "อนุมัติ")

	assert.ErrorIs(t, err, domain.ErrRequestAlreadyProcessed)
}

func TestLeaveService_Approve_SelfApproval(t *testing.T) {
	userID := domain.NewID()

	request := domain.NewLeaveRequest(
		userID, domain.LeaveTypeSick,
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		"ไม่สบาย",
	)

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	err := svc.Approve(context.Background(), request.ID, userID, "อนุมัติ")

	assert.ErrorIs(t, err, domain.ErrSelfApproval)
}

func TestLeaveService_Approve_NotPending(t *testing.T) {
	employeeID := domain.NewID()
	managerID := domain.NewID()

	// สร้างใบลาที่ถูกอนุมัติแล้ว
	request := domain.NewLeaveRequest(
		employeeID, domain.LeaveTypeSick,
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		"ไม่สบาย",
	)
	_ = request.Approve(managerID, "อนุมัติแล้ว")

	reviewerID := domain.NewID()

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	err := svc.Approve(context.Background(), request.ID, reviewerID, "อนุมัติอีกครั้ง")

	assert.ErrorIs(t, err, domain.ErrRequestNotPending)
}

func TestLeaveService_Reject_Success(t *testing.T) {
	employeeID := domain.NewID()
	managerID := domain.NewID()

	request := domain.NewLeaveRequest(
		employeeID, domain.LeaveTypePersonal,
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		"ธุระส่วนตัว",
	)

	var updatedRequest *domain.LeaveRequest

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
		updateWithStatusCheckFn: func(_ context.Context, r *domain.LeaveRequest, _ domain.LeaveStatus) error {
			updatedRequest = r
			return nil
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	err := svc.Reject(context.Background(), request.ID, managerID, "ช่วงเวลานี้มีงานเร่งด่วน")

	require.NoError(t, err)
	assert.Equal(t, domain.LeaveStatusRejected, updatedRequest.Status)
	assert.Equal(t, "ช่วงเวลานี้มีงานเร่งด่วน", updatedRequest.ReviewNote)
}

func TestLeaveService_GetMyRequests_Success(t *testing.T) {
	userID := domain.NewID()
	params := domain.NewPaginationParams(1, 10)
	expected := []domain.LeaveRequest{
		*domain.NewLeaveRequest(userID, domain.LeaveTypeSick,
			time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC), "ไม่สบาย"),
		*domain.NewLeaveRequest(userID, domain.LeaveTypeAnnual,
			time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC), "พักผ่อน"),
	}

	requestRepo := &mockLeaveRequestRepository{
		findByUserIDFn: func(_ context.Context, _ domain.ID, p domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error) {
			return domain.NewPaginatedResult(expected, 2, p), nil
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	result, err := svc.GetMyRequests(context.Background(), userID, params)

	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
	assert.Equal(t, 1, result.TotalPages)
}

func TestLeaveService_GetMyBalance_Success(t *testing.T) {
	userID := domain.NewID()
	expected := []domain.LeaveBalance{
		*domain.NewLeaveBalance(userID, domain.LeaveTypeSick, 30, 2026),
		*domain.NewLeaveBalance(userID, domain.LeaveTypeAnnual, 15, 2026),
		*domain.NewLeaveBalance(userID, domain.LeaveTypePersonal, 10, 2026),
	}

	balanceRepo := &mockLeaveBalanceRepository{
		findByUserIDFn: func(_ context.Context, _ domain.ID) ([]domain.LeaveBalance, error) {
			return expected, nil
		},
	}

	svc := NewLeaveService(&mockLeaveRequestRepository{}, balanceRepo)

	balances, err := svc.GetMyBalance(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, balances, 3)
}

func TestLeaveService_GetPendingRequests_Success(t *testing.T) {
	params := domain.NewPaginationParams(1, 10)
	expected := []domain.LeaveRequest{
		*domain.NewLeaveRequest(domain.NewID(), domain.LeaveTypeSick,
			time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC), "ไม่สบาย"),
	}

	requestRepo := &mockLeaveRequestRepository{
		findByStatusFn: func(_ context.Context, status domain.LeaveStatus, p domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error) {
			assert.Equal(t, domain.LeaveStatusPending, status)
			return domain.NewPaginatedResult(expected, 1, p), nil
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	result, err := svc.GetPendingRequests(context.Background(), params)

	require.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, 1, result.TotalPages)
}

func TestLeaveService_Reject_SelfRejection(t *testing.T) {
	userID := domain.NewID()

	request := domain.NewLeaveRequest(
		userID, domain.LeaveTypeSick,
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		"test",
	)

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
	}

	svc := NewLeaveService(requestRepo, &mockLeaveBalanceRepository{})

	err := svc.Reject(context.Background(), request.ID, userID, "note")

	assert.ErrorIs(t, err, domain.ErrSelfApproval)
}

// ─── Double Submit Protection Tests ─────────────────────────────────────

func TestLeaveService_Submit_DoubleSubmit_SecondRequestFails(t *testing.T) {
	// จำลองสถานการณ์: พนักงานมีวันลาเหลือ 1 วัน กดส่ง request ลาป่วย 2 ครั้งพร้อมกัน
	// request แรกจะจอง pending_days สำเร็จ
	// request ที่สอง → atomic ReservePending จะเห็นว่า used+pending+new > total → reject

	userID := domain.NewID()
	callCount := 0

	requestRepo := &mockLeaveRequestRepository{
		hasOverlapFn: func(_ context.Context, _ domain.ID, _, _ time.Time, _ *domain.ID) (bool, error) {
			return false, nil // วันที่ต่างกัน — ไม่ overlap
		},
	}
	balanceRepo := &mockLeaveBalanceRepository{
		reservePendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, _ float64) error {
			callCount++
			if callCount == 1 {
				return nil // request แรก: จองสำเร็จ
			}
			// request ที่สอง: atomic condition ล้มเหลว — used(0) + pending(1) + new(1) > total(1)
			return domain.ErrInsufficientBalance
		},
	}

	svc := NewLeaveService(requestRepo, balanceRepo)

	// request แรก — สำเร็จ
	req1, err := svc.Submit(context.Background(), userID, domain.LeaveTypeSick,
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		"ไม่สบาย ครั้งที่ 1",
	)
	require.NoError(t, err)
	assert.NotNil(t, req1)

	// request ที่สอง (double submit) — ถูก reject เพราะ pending_days ถูกจองไปแล้ว
	req2, err := svc.Submit(context.Background(), userID, domain.LeaveTypeSick,
		time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
		"ไม่สบาย ครั้งที่ 2",
	)
	assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
	assert.Nil(t, req2)
}

func TestLeaveService_Reject_ReleasesPendingDays(t *testing.T) {
	// ทดสอบว่าเมื่อปฏิเสธใบลา จะปล่อย pending_days กลับคืน
	employeeID := domain.NewID()
	managerID := domain.NewID()

	request := domain.NewLeaveRequest(
		employeeID, domain.LeaveTypeSick,
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		"ไม่สบาย",
	)

	var releasedDays float64

	requestRepo := &mockLeaveRequestRepository{
		findByIDFn: func(_ context.Context, _ domain.ID) (*domain.LeaveRequest, error) {
			return request, nil
		},
		updateWithStatusCheckFn: func(_ context.Context, _ *domain.LeaveRequest, _ domain.LeaveStatus) error {
			return nil
		},
	}
	balanceRepo := &mockLeaveBalanceRepository{
		releasePendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, days float64) error {
			releasedDays = days
			return nil
		},
	}

	svc := NewLeaveService(requestRepo, balanceRepo)

	err := svc.Reject(context.Background(), request.ID, managerID, "ไม่อนุมัติ")

	require.NoError(t, err)
	assert.Equal(t, float64(1), releasedDays) // ต้องปล่อย 1 วันกลับคืน
}

func TestLeaveService_Submit_RollbackOnCreateFailure(t *testing.T) {
	// ทดสอบว่าถ้าบันทึกใบลาล้มเหลว ต้อง rollback pending reservation
	var released bool

	requestRepo := &mockLeaveRequestRepository{
		hasOverlapFn: func(_ context.Context, _ domain.ID, _, _ time.Time, _ *domain.ID) (bool, error) {
			return false, nil
		},
		createFn: func(_ context.Context, _ *domain.LeaveRequest) error {
			return fmt.Errorf("database connection error") // จำลองว่า DB ล้มเหลว
		},
	}
	balanceRepo := &mockLeaveBalanceRepository{
		reservePendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, _ float64) error {
			return nil // จองสำเร็จ
		},
		releasePendingFn: func(_ context.Context, _ domain.ID, _ domain.LeaveType, _ int, _ float64) error {
			released = true // ต้องถูกเรียก rollback
			return nil
		},
	}

	svc := NewLeaveService(requestRepo, balanceRepo)

	_, err := svc.Submit(context.Background(), domain.NewID(), domain.LeaveTypeSick,
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		"ไม่สบาย",
	)

	assert.Error(t, err)
	assert.True(t, released, "ต้อง rollback pending reservation เมื่อบันทึกใบลาล้มเหลว")
}
