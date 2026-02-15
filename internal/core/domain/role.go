package domain

type Role string

const (
	RoleEmployee Role = "employee" // คือพนักงานทั่วไป — สามารถยื่นใบลาและดูประวัติการลาได้
	RoleManager  Role = "manager"  // คือผู้จัดการ — สามารถอนุมัติหรือปฏิเสธใบลาได้
)

func AllRoles() []Role {
	return []Role{RoleEmployee, RoleManager}
}

func (r Role) IsValid() bool {
	switch r {
	case RoleEmployee, RoleManager:
		return true
	default:
		return false
	}
}

func (r Role) String() string {
	return string(r)
}
