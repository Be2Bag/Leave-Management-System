package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github/be2bag/leave-management-system/internal/core/domain"
)

// ─── Leave Balance Tests ────────────────────────────────────────────────
// ทดสอบ business logic ของยอดวันลา
// ─────────────────────────────────────────────────────────────────────────

func TestLeaveBalance_RemainingDays(t *testing.T) {
	// ทดสอบการคำนวณวันลาคงเหลือ
	balance := domain.NewLeaveBalance(domain.NewID(), domain.LeaveTypeSick, 30, 2026)

	assert.Equal(t, 30.0, balance.RemainingDays(), "ยอดคงเหลือเริ่มต้นต้องเท่ากับ total")

	balance.UsedDays = 10
	assert.Equal(t, 20.0, balance.RemainingDays(), "ยอดคงเหลือต้องลดลงตามที่ใช้ไป")
}

func TestLeaveBalance_PendingDays_AffectsAvailability(t *testing.T) {
	// ทดสอบว่า PendingDays ถูกนับรวมในการตรวจสอบยอดคงเหลือ
	balance := domain.NewLeaveBalance(domain.NewID(), domain.LeaveTypeSick, 10, 2026)
	balance.UsedDays = 5
	balance.PendingDays = 4

	// คงเหลือจริง = 10 - 5 - 4 = 1
	assert.Equal(t, 1.0, balance.RemainingDays(), "remaining ต้องหัก pending ด้วย")
}

// ─── Leave Request Tests ────────────────────────────────────────────────
// ทดสอบ workflow ของคำขอลา
// ─────────────────────────────────────────────────────────────────────────

func TestLeaveRequest_NewLeaveRequest(t *testing.T) {
	userID := domain.NewID()
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)

	request := domain.NewLeaveRequest(userID, domain.LeaveTypeAnnual, start, end, "ลาพักร้อน")

	assert.Equal(t, domain.LeaveStatusPending, request.Status, "สถานะเริ่มต้นต้องเป็น pending")
	assert.Equal(t, 3.0, request.TotalDays, "จำนวนวันต้องเป็น 3")
}

func TestLeaveRequest_Approve_Success(t *testing.T) {
	userID := domain.NewID()
	reviewerID := domain.NewID()
	request := domain.NewLeaveRequest(userID, domain.LeaveTypeSick,
		time.Now(), time.Now(), "ป่วย")

	err := request.Approve(reviewerID, "อนุมัติแล้ว")

	assert.NoError(t, err)
	assert.Equal(t, domain.LeaveStatusApproved, request.Status, "สถานะต้องเป็น approved")
	assert.NotNil(t, request.ReviewerID, "reviewer_id ต้องไม่เป็น nil")
	assert.NotNil(t, request.ReviewedAt, "reviewed_at ต้องไม่เป็น nil")
}

func TestLeaveRequest_Approve_NotPending(t *testing.T) {
	// ทดสอบว่าอนุมัติใบลาที่ไม่ใช่ pending จะ error
	request := domain.NewLeaveRequest(domain.NewID(), domain.LeaveTypeSick,
		time.Now(), time.Now(), "ป่วย")

	// อนุมัติครั้งแรก
	_ = request.Approve(domain.NewID(), "อนุมัติ")

	// พยายามอนุมัติอีกครั้ง — ต้อง error
	err := request.Approve(domain.NewID(), "อนุมัติอีกครั้ง")

	assert.ErrorIs(t, err, domain.ErrRequestNotPending, "ต้อง error เมื่ออนุมัติซ้ำ")
}

func TestLeaveRequest_Reject_Success(t *testing.T) {
	reviewerID := domain.NewID()
	request := domain.NewLeaveRequest(domain.NewID(), domain.LeaveTypeSick,
		time.Now(), time.Now(), "ป่วย")

	err := request.Reject(reviewerID, "ไม่อนุมัติ")

	assert.NoError(t, err)
	assert.Equal(t, domain.LeaveStatusRejected, request.Status, "สถานะต้องเป็น rejected")
}

func TestLeaveRequest_Reject_NotPending(t *testing.T) {
	request := domain.NewLeaveRequest(domain.NewID(), domain.LeaveTypeSick,
		time.Now(), time.Now(), "ป่วย")
	_ = request.Reject(domain.NewID(), "ปฏิเสธ")

	err := request.Reject(domain.NewID(), "ปฏิเสธซ้ำ")
	assert.ErrorIs(t, err, domain.ErrRequestNotPending, "ต้อง error เมื่อปฏิเสธซ้ำ")
}

// ─── CalculateLeaveDays Tests ───────────────────────────────────────────
// ทดสอบการคำนวณจำนวนวันลา
// ─────────────────────────────────────────────────────────────────────────

func TestCalculateLeaveDays(t *testing.T) {
	tests := []struct {
		start    time.Time
		end      time.Time
		name     string
		expected float64
	}{
		{
			name:     "ลา 1 วัน (วันเดียว)",
			start:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "ลา 3 วัน (1-3 มีนาคม)",
			start:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
			expected: 3,
		},
		{
			name:     "ลา 5 วัน (1-5 มกราคม)",
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			expected: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.CalculateLeaveDays(tc.start, tc.end)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ─── Role & LeaveType Tests ─────────────────────────────────────────────

func TestRole_IsValid(t *testing.T) {
	assert.True(t, domain.RoleEmployee.IsValid(), "employee ต้อง valid")
	assert.True(t, domain.RoleManager.IsValid(), "manager ต้อง valid")
	assert.False(t, domain.Role("admin").IsValid(), "admin ต้อง invalid")
}

func TestLeaveType_IsValid(t *testing.T) {
	assert.True(t, domain.LeaveTypeSick.IsValid())
	assert.True(t, domain.LeaveTypeAnnual.IsValid())
	assert.True(t, domain.LeaveTypePersonal.IsValid())
	assert.False(t, domain.LeaveType("maternity").IsValid())
}

func TestLeaveStatus_IsValid(t *testing.T) {
	assert.True(t, domain.LeaveStatusPending.IsValid())
	assert.True(t, domain.LeaveStatusApproved.IsValid())
	assert.True(t, domain.LeaveStatusRejected.IsValid())
	assert.False(t, domain.LeaveStatus("cancelled").IsValid())
}
