package domain

import "time"

// LeaveRequest คำขอลาของพนักงาน
type LeaveRequest struct {
	StartDate  time.Time   `json:"start_date"            bson:"start_date"`            // วันเริ่มต้นลา
	EndDate    time.Time   `json:"end_date"              bson:"end_date"`              // วันสิ้นสุดลา
	CreatedAt  time.Time   `json:"created_at"            bson:"created_at"`            // วันที่ยื่นใบลา
	UpdatedAt  time.Time   `json:"updated_at"            bson:"updated_at"`            // วันที่แก้ไขล่าสุด
	ReviewedAt *time.Time  `json:"reviewed_at,omitempty" bson:"reviewed_at,omitempty"` // วันที่อนุมัติ/ปฏิเสธ
	ReviewerID *ID         `json:"reviewer_id,omitempty" bson:"reviewer_id,omitempty"` // รหัสผู้อนุมัติ
	LeaveType  LeaveType   `json:"leave_type"            bson:"leave_type"`            // ประเภทการลา
	Reason     string      `json:"reason"                bson:"reason"`                // เหตุผลการลา
	ReviewNote string      `json:"review_note,omitempty" bson:"review_note,omitempty"` // หมายเหตุจากผู้อนุมัติ
	Status     LeaveStatus `json:"status"                bson:"status"`                // สถานะใบลา
	ID         ID          `json:"id"                    bson:"_id"`                   // รหัสใบลา (UUID)
	UserID     ID          `json:"user_id"               bson:"user_id"`               // รหัสพนักงานที่ยื่นใบลา
	TotalDays  float64     `json:"total_days"            bson:"total_days"`            // จำนวนวันลาทั้งหมด
}

func NewLeaveRequest(userID ID, leaveType LeaveType, startDate, endDate time.Time, reason string) *LeaveRequest {
	now := time.Now()
	totalDays := CalculateLeaveDays(startDate, endDate)
	return &LeaveRequest{
		ID:        NewID(),
		UserID:    userID,
		LeaveType: leaveType,
		StartDate: startDate,
		EndDate:   endDate,
		TotalDays: totalDays,
		Reason:    reason,
		Status:    LeaveStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Approve อนุมัติใบลา — ทำได้เฉพาะใบลาที่อยู่ในสถานะ Pending เท่านั้น
func (r *LeaveRequest) Approve(reviewerID ID, note string) error {
	if r.Status != LeaveStatusPending {
		return ErrRequestNotPending
	}
	now := time.Now()
	r.Status = LeaveStatusApproved
	r.ReviewerID = &reviewerID
	r.ReviewNote = note
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}

// Reject ปฏิเสธใบลา — ทำได้เฉพาะใบลาที่อยู่ในสถานะ Pending เท่านั้น
func (r *LeaveRequest) Reject(reviewerID ID, note string) error {
	if r.Status != LeaveStatusPending {
		return ErrRequestNotPending
	}
	now := time.Now()
	r.Status = LeaveStatusRejected
	r.ReviewerID = &reviewerID
	r.ReviewNote = note
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}

func (r *LeaveRequest) IsPending() bool {
	return r.Status == LeaveStatusPending
}

// CalculateLeaveDays คำนวณจำนวนวันลาจากวันเริ่มต้นถึงวันสิ้นสุด (นับรวมวันเริ่มต้น)
func CalculateLeaveDays(startDate, endDate time.Time) float64 {
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
	days := end.Sub(start).Hours() / 24
	return days + 1 // +1 เพราะนับรวมวันเริ่มต้นด้วย
}
