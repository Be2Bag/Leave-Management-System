package domain

import "time"

// LeaveBalance ยอดคงเหลือวันลาของพนักงาน
type LeaveBalance struct {
	CreatedAt   time.Time `json:"created_at"   bson:"created_at"`   // วันที่สร้าง
	UpdatedAt   time.Time `json:"updated_at"   bson:"updated_at"`   // วันที่แก้ไขล่าสุด
	LeaveType   LeaveType `json:"leave_type"   bson:"leave_type"`   // ประเภทการลา
	ID          ID        `json:"id"           bson:"_id"`          // รหัสยอดวันลา (UUID)
	UserID      ID        `json:"user_id"      bson:"user_id"`      // รหัสพนักงานเจ้าของยอดวันลา
	TotalDays   float64   `json:"total_days"   bson:"total_days"`   // จำนวนวันลาทั้งหมดที่ได้รับ
	UsedDays    float64   `json:"used_days"    bson:"used_days"`    // จำนวนวันลาที่ใช้ไปแล้ว (อนุมัติแล้ว)
	PendingDays float64   `json:"pending_days" bson:"pending_days"` // จำนวนวันลาที่จองไว้ (รอการอนุมัติ)
	Year        int       `json:"year"         bson:"year"`         // ปีที่ยอดวันลานี้ใช้ได้
}

// RemainingDays คำนวณจำนวนวันลาคงเหลือ (หักทั้งที่ใช้แล้วและที่จองไว้)
func (b *LeaveBalance) RemainingDays() float64 {
	return b.TotalDays - b.UsedDays - b.PendingDays
}

// AvailableDays คำนวณจำนวนวันลาที่ยังสามารถขอได้ (รวมที่จอง pending ด้วย)
func (b *LeaveBalance) AvailableDays() float64 {
	return b.TotalDays - b.UsedDays - b.PendingDays
}

// HasSufficientBalance ตรวจสอบว่ามีวันลาเพียงพอสำหรับจำนวนวันที่ต้องการหรือไม่
func (b *LeaveBalance) HasSufficientBalance(days float64) bool {
	return b.AvailableDays() >= days
}

// Deduct หักวันลาจากยอดคงเหลือ — คืน error หากวันลาไม่เพียงพอ
func (b *LeaveBalance) Deduct(days float64) error {
	if !b.HasSufficientBalance(days) {
		return ErrInsufficientBalance
	}
	b.UsedDays += days
	b.UpdatedAt = time.Now()
	return nil
}

// Restore คืนวันลากลับเข้ายอดคงเหลือ (ใช้เมื่อใบลาถูกยกเลิก)
func (b *LeaveBalance) Restore(days float64) {
	b.UsedDays -= days
	if b.UsedDays < 0 {
		b.UsedDays = 0
	}
	b.UpdatedAt = time.Now()
}

func NewLeaveBalance(userID ID, leaveType LeaveType, totalDays float64, year int) *LeaveBalance {
	now := time.Now()
	return &LeaveBalance{
		ID:        NewID(),
		UserID:    userID,
		LeaveType: leaveType,
		TotalDays: totalDays,
		UsedDays:  0,
		Year:      year,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
