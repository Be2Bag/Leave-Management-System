package domain

type TokenClaims struct {
	Email  string // อีเมลผู้ใช้
	Role   Role   // บทบาทของผู้ใช้ (employee/manager)
	UserID ID     // รหัสผู้ใช้ (UUID)
}
