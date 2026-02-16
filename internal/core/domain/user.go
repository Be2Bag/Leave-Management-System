package domain

import "time"

// User ข้อมูลผู้ใช้งานในระบบ
type User struct {
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`    // วันที่สร้าง
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`    // วันที่แก้ไขล่าสุด
	FirstName    string    `json:"first_name" bson:"first_name"`    // ชื่อจริง
	LastName     string    `json:"last_name"  bson:"last_name"`     // นามสกุล
	FullName     string    `json:"full_name"  bson:"full_name"`     // ชื่อเต็ม (first + last)
	Email        string    `json:"email"      bson:"email"`         // อีเมล (unique)
	PasswordHash string    `json:"-"          bson:"password_hash"` // รหัสผ่านที่เข้ารหัสแล้ว (ไม่ส่งกลับใน JSON)
	Role         Role      `json:"role"       bson:"role"`          // บทบาท (employee/manager)
	ID           ID        `json:"user_id"    bson:"_id"`           // รหัสผู้ใช้ (UUID) — ใช้เป็น primary key
}

func NewUser(firstName, lastName, email, passwordHash string, role Role) *User {
	now := time.Now()
	return &User{
		ID:           NewID(),
		FirstName:    firstName,
		LastName:     lastName,
		FullName:     firstName + " " + lastName,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
