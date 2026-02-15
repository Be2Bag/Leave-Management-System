package dto

import (
	"time"

	"github/be2bag/leave-management-system/internal/core/domain"
)

type SubmitLeaveRequest struct {
	LeaveType string `json:"leave_type" validate:"required,oneof=sick_leave annual_leave personal_leave"` // ประเภทการลา
	StartDate string `json:"start_date" validate:"required"`                                              // วันเริ่มต้น (YYYY-MM-DD)
	EndDate   string `json:"end_date"   validate:"required"`                                              // วันสิ้นสุด (YYYY-MM-DD)
	Reason    string `json:"reason"     validate:"required,min=5,max=500"`                                // เหตุผลการลา
}

type ReviewLeaveRequest struct {
	Note string `json:"note" validate:"max=500"` // หมายเหตุจากผู้อนุมัติ (ไม่บังคับ)
}

type LeaveRequestResponse struct {
	ID         string  `json:"id"`                    // รหัสใบลา
	UserID     string  `json:"user_id"`               // รหัสพนักงาน
	LeaveType  string  `json:"leave_type"`            // ประเภทการลา
	StartDate  string  `json:"start_date"`            // วันเริ่มต้น
	EndDate    string  `json:"end_date"`              // วันสิ้นสุด
	Reason     string  `json:"reason"`                // เหตุผลการลา
	Status     string  `json:"status"`                // สถานะ (pending/approved/rejected)
	ReviewerID string  `json:"reviewer_id,omitempty"` // รหัสผู้อนุมัติ
	ReviewNote string  `json:"review_note,omitempty"` // หมายเหตุจากผู้อนุมัติ
	ReviewedAt string  `json:"reviewed_at,omitempty"` // วันที่อนุมัติ/ปฏิเสธ
	CreatedAt  string  `json:"created_at"`            // วันที่ยื่นใบลา
	UpdatedAt  string  `json:"updated_at"`            // วันที่แก้ไขล่าสุด
	TotalDays  float64 `json:"total_days"`            // จำนวนวันลาทั้งหมด
}

type LeaveBalanceResponse struct {
	ID            string  `json:"id"`             // รหัสยอดวันลา
	LeaveType     string  `json:"leave_type"`     // ประเภทการลา
	TotalDays     float64 `json:"total_days"`     // วันลาทั้งหมด
	UsedDays      float64 `json:"used_days"`      // วันลาที่ใช้ไป
	RemainingDays float64 `json:"remaining_days"` // วันลาคงเหลือ
	Year          int     `json:"year"`           // ปี
}

func ToLeaveRequestResponse(r *domain.LeaveRequest) LeaveRequestResponse {
	resp := LeaveRequestResponse{
		ID:         r.ID.String(),
		UserID:     r.UserID.String(),
		LeaveType:  string(r.LeaveType),
		StartDate:  r.StartDate.Format("2006-01-02"),
		EndDate:    r.EndDate.Format("2006-01-02"),
		TotalDays:  r.TotalDays,
		Reason:     r.Reason,
		Status:     string(r.Status),
		ReviewNote: r.ReviewNote,
		CreatedAt:  r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  r.UpdatedAt.Format(time.RFC3339),
	}

	if r.ReviewerID != nil {
		resp.ReviewerID = r.ReviewerID.String()
	}
	if r.ReviewedAt != nil {
		resp.ReviewedAt = r.ReviewedAt.Format(time.RFC3339)
	}

	return resp
}

func ToLeaveRequestResponses(requests []domain.LeaveRequest) []LeaveRequestResponse {
	responses := make([]LeaveRequestResponse, 0, len(requests))
	for i := range requests {
		responses = append(responses, ToLeaveRequestResponse(&requests[i]))
	}
	return responses
}

func ToLeaveBalanceResponse(b *domain.LeaveBalance) LeaveBalanceResponse {
	return LeaveBalanceResponse{
		ID:            b.ID.String(),
		LeaveType:     string(b.LeaveType),
		TotalDays:     b.TotalDays,
		UsedDays:      b.UsedDays,
		RemainingDays: b.RemainingDays(),
		Year:          b.Year,
	}
}

func ToLeaveBalanceResponses(balances []domain.LeaveBalance) []LeaveBalanceResponse {
	responses := make([]LeaveBalanceResponse, 0, len(balances))
	for i := range balances {
		responses = append(responses, ToLeaveBalanceResponse(&balances[i]))
	}
	return responses
}
