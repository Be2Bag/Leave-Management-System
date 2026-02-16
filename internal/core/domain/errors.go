package domain

import "errors"

var (
	// ─── User Errors ────────────────────────────────────────────────

	ErrUserNotFound       = errors.New("ไม่พบผู้ใช้ในระบบ")
	ErrInvalidCredentials = errors.New("อีเมลหรือรหัสผ่านไม่ถูกต้อง")

	// ─── Leave Errors ───────────────────────────────────────────────

	ErrInvalidLeaveType     = errors.New("ประเภทการลาไม่ถูกต้อง")
	ErrInsufficientBalance  = errors.New("วันลาคงเหลือไม่เพียงพอ")
	ErrOverlappingLeave     = errors.New("วันลาซ้ำซ้อนกับใบลาที่มีอยู่แล้ว")
	ErrInvalidDateRange     = errors.New("ช่วงวันที่ไม่ถูกต้อง: วันสิ้นสุดต้องไม่ก่อนวันเริ่มต้น")
	ErrLeaveBalanceNotFound = errors.New("ไม่พบข้อมูลยอดวันลาสำหรับประเภทและปีที่ระบุ")

	// ─── Leave Request Errors ───────────────────────────────────────

	ErrRequestNotFound         = errors.New("ไม่พบคำขอลา")
	ErrRequestNotPending       = errors.New("ใบลาไม่อยู่ในสถานะรอดำเนินการ")
	ErrRequestAlreadyProcessed = errors.New("ใบลาถูกดำเนินการไปแล้ว")
	ErrSelfApproval            = errors.New("ไม่สามารถอนุมัติหรือปฏิเสธใบลาของตนเองได้")

	// ─── Auth Errors ────────────────────────────────────────────────

	ErrUnauthorized = errors.New("ไม่มีสิทธิ์เข้าถึง")
)
