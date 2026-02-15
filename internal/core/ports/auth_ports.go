package ports

import (
	"context"

	"github/be2bag/leave-management-system/internal/core/domain"
)

type AuthService interface {
	// Login เข้าสู่ระบบ — คืน JWT token และข้อมูลผู้ใช้
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
}

type TokenService interface {
	// GenerateToken สร้าง JWT token จากข้อมูลผู้ใช้
	GenerateToken(user *domain.User) (string, error)
	// ValidateToken ตรวจสอบและถอดรหัส JWT token — คืนข้อมูลผู้ใช้จาก token
	ValidateToken(tokenString string) (*domain.TokenClaims, error)
}
