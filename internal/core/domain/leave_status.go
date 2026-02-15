package domain

type LeaveStatus string // สถานะของใบลา

const (
	LeaveStatusPending  LeaveStatus = "pending"  // ใบลารอการอนุมัติจากผู้จัดการ
	LeaveStatusApproved LeaveStatus = "approved" // ใบลาได้รับการอนุมัติแล้ว — หักวันลาจากยอดคงเหลือ
	LeaveStatusRejected LeaveStatus = "rejected" // ใบลาถูกปฏิเสธ — ยอดวันลาไม่เปลี่ยนแปลง
)

func (s LeaveStatus) IsValid() bool {
	switch s {
	case LeaveStatusPending, LeaveStatusApproved, LeaveStatusRejected:
		return true
	default:
		return false
	}
}

func (s LeaveStatus) String() string {
	return string(s)
}
