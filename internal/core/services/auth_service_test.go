package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github/be2bag/leave-management-system/internal/core/domain"
)

func TestAuthService_Login_Success(t *testing.T) {
	// สร้างรหัสผ่าน hash สำหรับ "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	testUser := domain.NewUser("สมชาย", "ใจดี", "somchai@company.com", string(hashedPassword), domain.RoleEmployee)

	userRepo := &mockUserRepository{
		findByEmailFn: func(_ context.Context, email string) (*domain.User, error) {
			if email == "somchai@company.com" {
				return testUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	tokenSvc := &mockTokenService{
		generateFn: func(_ *domain.User) (string, error) {
			return "jwt-token-123", nil
		},
	}

	svc := NewAuthService(userRepo, tokenSvc)

	// Act
	token, user, err := svc.Login(context.Background(), "somchai@company.com", "password123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "jwt-token-123", token)
	assert.Equal(t, "สมชาย", user.FirstName)
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	userRepo := &mockUserRepository{
		findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := NewAuthService(userRepo, &mockTokenService{})

	_, _, err := svc.Login(context.Background(), "nonexistent@company.com", "password123")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.MinCost)
	testUser := domain.NewUser("Test", "User", "test@test.com", string(hashedPassword), domain.RoleEmployee)

	userRepo := &mockUserRepository{
		findByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return testUser, nil
		},
	}
	svc := NewAuthService(userRepo, &mockTokenService{})

	_, _, err := svc.Login(context.Background(), "test@test.com", "wrong_password")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}
