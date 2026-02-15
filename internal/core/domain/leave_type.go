package domain

type LeaveType string // ประเภทการลา

const (
	LeaveTypeSick     LeaveType = "sick_leave"     // ลาป่วย
	LeaveTypeAnnual   LeaveType = "annual_leave"   // ลาพักร้อน
	LeaveTypePersonal LeaveType = "personal_leave" // ลากิจ
)

func AllLeaveTypes() []LeaveType {
	return []LeaveType{LeaveTypeSick, LeaveTypeAnnual, LeaveTypePersonal}
}

func (t LeaveType) IsValid() bool {
	switch t {
	case LeaveTypeSick, LeaveTypeAnnual, LeaveTypePersonal:
		return true
	default:
		return false
	}
}

func (t LeaveType) String() string {
	return string(t)
}

func (t LeaveType) DisplayName() string {
	switch t {
	case LeaveTypeSick:
		return "ลาป่วย"
	case LeaveTypeAnnual:
		return "ลาพักร้อน"
	case LeaveTypePersonal:
		return "ลากิจ"
	default:
		return string(t)
	}
}

// DefaultQuota คืนจำนวนวันลาเริ่มต้นของแต่ละประเภท (วันต่อปี)
func (t LeaveType) DefaultQuota() float64 {
	switch t {
	case LeaveTypeSick:
		return 30
	case LeaveTypeAnnual:
		return 15
	case LeaveTypePersonal:
		return 10
	default:
		return 0
	}
}
