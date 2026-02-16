package domain

type LeaveType string // ประเภทการลา

const (
	LeaveTypeSick     LeaveType = "sick_leave"     // ลาป่วย
	LeaveTypeAnnual   LeaveType = "annual_leave"   // ลาพักร้อน
	LeaveTypePersonal LeaveType = "personal_leave" // ลากิจ
)

func (t LeaveType) IsValid() bool {
	switch t {
	case LeaveTypeSick, LeaveTypeAnnual, LeaveTypePersonal:
		return true
	default:
		return false
	}
}
