package services

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
)

type authService struct {
	userRepo     ports.UserRepository
	tokenService ports.TokenService
}

// NewAuthService สร้าง AuthService instance
func NewAuthService(
	userRepo ports.UserRepository,
	tokenService ports.TokenService,
) ports.AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// Login เข้าสู่ระบบ — ตรวจสอบอีเมลและรหัสผ่าน แล้วสร้าง JWT token
func (s *authService) Login(
	ctx context.Context,
	email, password string,
) (string, *domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("สร้าง token ล้มเหลว: %w", err)
	}

	return token, user, nil
}
