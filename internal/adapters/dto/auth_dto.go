package dto

import (
	"time"

	"github/be2bag/leave-management-system/internal/core/domain"
)

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"` // อีเมล
	Password string `json:"password" validate:"required"`       // รหัสผ่าน
}

type AuthResponse struct {
	Token string       `json:"token"` // JWT access token
	User  UserResponse `json:"user"`  // ข้อมูลผู้ใช้
}

type UserResponse struct {
	ID        string `json:"user_id"`    // รหัสผู้ใช้ (UUID)
	FirstName string `json:"first_name"` // ชื่อจริง
	LastName  string `json:"last_name"`  // นามสกุล
	FullName  string `json:"full_name"`  // ชื่อเต็ม
	Email     string `json:"email"`      // อีเมล
	Role      string `json:"role"`       // บทบาท
	CreatedAt string `json:"created_at"` // วันที่สร้างบัญชี
}

func ToUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FullName:  user.FullName,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
