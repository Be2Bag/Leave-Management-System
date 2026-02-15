package services

import (
	"context"
	"time"

	"github/be2bag/leave-management-system/internal/core/domain"
)

// ─── Mock Repository Implementations ────────────────────────────────────
// Mock สำหรับทดสอบ service layer — จำลองพฤติกรรมของ repository
// ─────────────────────────────────────────────────────────────────────────

// mockUserRepository จำลอง UserRepository สำหรับทดสอบ
type mockUserRepository struct {
	findByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, domain.ErrUserNotFound
}

// mockLeaveBalanceRepository จำลอง LeaveBalanceRepository สำหรับทดสอบ
type mockLeaveBalanceRepository struct {
	findByUserIDFn   func(ctx context.Context, userID domain.ID) ([]domain.LeaveBalance, error)
	reservePendingFn func(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error
	confirmPendingFn func(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error
	releasePendingFn func(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error
}

func (m *mockLeaveBalanceRepository) FindByUserID(ctx context.Context, userID domain.ID) ([]domain.LeaveBalance, error) {
	if m.findByUserIDFn != nil {
		return m.findByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockLeaveBalanceRepository) ReservePending(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error {
	if m.reservePendingFn != nil {
		return m.reservePendingFn(ctx, userID, leaveType, year, days)
	}
	return nil
}

func (m *mockLeaveBalanceRepository) ConfirmPending(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error {
	if m.confirmPendingFn != nil {
		return m.confirmPendingFn(ctx, userID, leaveType, year, days)
	}
	return nil
}

func (m *mockLeaveBalanceRepository) ReleasePending(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error {
	if m.releasePendingFn != nil {
		return m.releasePendingFn(ctx, userID, leaveType, year, days)
	}
	return nil
}

// mockLeaveRequestRepository จำลอง LeaveRequestRepository สำหรับทดสอบ
type mockLeaveRequestRepository struct {
	createFn                func(ctx context.Context, request *domain.LeaveRequest) error
	findByIDFn              func(ctx context.Context, id domain.ID) (*domain.LeaveRequest, error)
	findByUserIDFn          func(ctx context.Context, userID domain.ID, params domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error)
	findByStatusFn          func(ctx context.Context, status domain.LeaveStatus, params domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error)
	updateFn                func(ctx context.Context, request *domain.LeaveRequest) error
	updateWithStatusCheckFn func(ctx context.Context, request *domain.LeaveRequest, expectedStatus domain.LeaveStatus) error
	hasOverlapFn            func(ctx context.Context, userID domain.ID, startDate, endDate time.Time, excludeID *domain.ID) (bool, error)
}

func (m *mockLeaveRequestRepository) Create(ctx context.Context, request *domain.LeaveRequest) error {
	if m.createFn != nil {
		return m.createFn(ctx, request)
	}
	return nil
}

func (m *mockLeaveRequestRepository) FindByID(ctx context.Context, id domain.ID) (*domain.LeaveRequest, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrRequestNotFound
}

func (m *mockLeaveRequestRepository) FindByUserID(
	ctx context.Context,
	userID domain.ID,
	params domain.PaginationParams,
) (*domain.PaginatedResult[domain.LeaveRequest], error) {
	if m.findByUserIDFn != nil {
		return m.findByUserIDFn(ctx, userID, params)
	}
	return domain.NewPaginatedResult([]domain.LeaveRequest{}, 0, params), nil
}

func (m *mockLeaveRequestRepository) FindByStatus(
	ctx context.Context,
	status domain.LeaveStatus,
	params domain.PaginationParams,
) (*domain.PaginatedResult[domain.LeaveRequest], error) {
	if m.findByStatusFn != nil {
		return m.findByStatusFn(ctx, status, params)
	}
	return domain.NewPaginatedResult([]domain.LeaveRequest{}, 0, params), nil
}

func (m *mockLeaveRequestRepository) Update(ctx context.Context, request *domain.LeaveRequest) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, request)
	}
	return nil
}

func (m *mockLeaveRequestRepository) UpdateWithStatusCheck(ctx context.Context, request *domain.LeaveRequest, expectedStatus domain.LeaveStatus) error {
	if m.updateWithStatusCheckFn != nil {
		return m.updateWithStatusCheckFn(ctx, request, expectedStatus)
	}
	return nil
}

func (m *mockLeaveRequestRepository) HasOverlap(
	ctx context.Context,
	userID domain.ID,
	startDate, endDate time.Time,
	excludeID *domain.ID,
) (bool, error) {
	if m.hasOverlapFn != nil {
		return m.hasOverlapFn(ctx, userID, startDate, endDate, excludeID)
	}
	return false, nil
}

// mockTokenService จำลอง TokenService สำหรับทดสอบ
type mockTokenService struct {
	generateFn func(user *domain.User) (string, error)
	validateFn func(tokenString string) (*domain.TokenClaims, error)
}

func (m *mockTokenService) GenerateToken(user *domain.User) (string, error) {
	if m.generateFn != nil {
		return m.generateFn(user)
	}
	return "mock-token", nil
}

func (m *mockTokenService) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	if m.validateFn != nil {
		return m.validateFn(tokenString)
	}
	return nil, domain.ErrUnauthorized
}
