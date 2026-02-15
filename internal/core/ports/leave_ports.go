package ports

import (
	"context"
	"time"

	"github/be2bag/leave-management-system/internal/core/domain"
)

type LeaveService interface {
	// Submit ยื่นใบลาใหม่ — ตรวจสอบ overlap และ balance ก่อนสร้าง
	Submit(ctx context.Context, userID domain.ID, leaveType domain.LeaveType,
		startDate, endDate time.Time, reason string) (*domain.LeaveRequest, error)
	// GetMyRequests ดูประวัติใบลาทั้งหมดของตนเอง (รองรับ pagination)
	GetMyRequests(ctx context.Context, userID domain.ID, params domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error)
	// GetMyBalance ดูยอดวันลาคงเหลือของตนเอง
	GetMyBalance(ctx context.Context, userID domain.ID) ([]domain.LeaveBalance, error)
	// GetPendingRequests ดูใบลาที่รอการอนุมัติ (สำหรับผู้จัดการ, รองรับ pagination)
	GetPendingRequests(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error)
	// Approve อนุมัติใบลา — หักยอดวันลาของพนักงาน
	Approve(ctx context.Context, requestID, reviewerID domain.ID, note string) error
	// Reject ปฏิเสธใบลา — ยอดวันลาไม่เปลี่ยนแปลง
	Reject(ctx context.Context, requestID, reviewerID domain.ID, note string) error
}

type LeaveBalanceRepository interface {
	// FindByUserID ค้นหายอดวันลาทั้งหมดของผู้ใช้
	FindByUserID(ctx context.Context, userID domain.ID) ([]domain.LeaveBalance, error)
	// ReservePending จองวันลาแบบ atomic — เพิ่ม pending_days เฉพาะเมื่อยอดเพียงพอ
	ReservePending(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error
	// ConfirmPending ยืนยันวันลาแบบ atomic — ย้ายจาก pending_days ไป used_days (ใช้ตอนอนุมัติ)
	ConfirmPending(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error
	// ReleasePending ปล่อยวันลาที่จองไว้แบบ atomic — ลด pending_days (ใช้ตอนปฏิเสธ)
	ReleasePending(ctx context.Context, userID domain.ID, leaveType domain.LeaveType, year int, days float64) error
}

type LeaveRequestRepository interface {
	// Create สร้างคำขอลาใหม่
	Create(ctx context.Context, request *domain.LeaveRequest) error
	// FindByID ค้นหาคำขอลาจากรหัส
	FindByID(ctx context.Context, id domain.ID) (*domain.LeaveRequest, error)
	// FindByUserID ค้นหาคำขอลาทั้งหมดของผู้ใช้ (รองรับ pagination)
	FindByUserID(ctx context.Context, userID domain.ID, params domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error)
	// FindByStatus ค้นหาคำขอลาตามสถานะ (เช่น pending สำหรับผู้จัดการ, รองรับ pagination)
	FindByStatus(ctx context.Context, status domain.LeaveStatus, params domain.PaginationParams) (*domain.PaginatedResult[domain.LeaveRequest], error)
	// Update อัปเดตคำขอลา (เช่น เปลี่ยนสถานะเป็น approved/rejected)
	Update(ctx context.Context, request *domain.LeaveRequest) error
	// UpdateWithStatusCheck อัปเดตคำขอลาแบบ atomic
	UpdateWithStatusCheck(ctx context.Context, request *domain.LeaveRequest, expectedStatus domain.LeaveStatus) error
	// HasOverlap ตรวจสอบว่ามีคำขอลาที่ซ้ำซ้อนกับช่วงวันที่ที่ระบุหรือไม่
	HasOverlap(ctx context.Context, userID domain.ID, startDate, endDate time.Time, excludeID *domain.ID) (bool, error)
}
