package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
)

type tokenService struct {
	secretKey   []byte
	expireHours int
}

func NewTokenService(secretKey string, expireHours int) ports.TokenService {
	return &tokenService{
		secretKey:   []byte(secretKey),
		expireHours: expireHours,
	}
}

// GenerateToken สร้าง JWT token จากข้อมูลผู้ใช้
func (s *tokenService) GenerateToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),                   // รหัสผู้ใช้
		"email":   user.Email,                         // อีเมล
		"role":    string(user.Role),                  // บทบาท
		"exp":     now.Add(s.expireDuration()).Unix(), // เวลาหมดอายุ
		"iat":     now.Unix(),                         // เวลาที่สร้าง
		"nbf":     now.Unix(),                         // ใช้ได้ตั้งแต่เวลานี้ (Not Before)
		"jti":     domain.NewID().String(),            // รหัสเฉพาะของ token (ใช้ revoke ได้ในอนาคต)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("สร้าง token ล้มเหลว: %w", err)
	}

	return signedToken, nil
}

// ValidateToken ตรวจสอบและถอดรหัส JWT token
func (s *tokenService) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	return s.extractClaims(claims)
}

// keyFunc ฟังก์ชันตรวจสอบ signing method และคืน secret key
func (s *tokenService) keyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return s.secretKey, nil
}

// extractClaims ดึงข้อมูลผู้ใช้จาก JWT claims
func (s *tokenService) extractClaims(claims jwt.MapClaims) (*domain.TokenClaims, error) {
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	userID, err := domain.ParseID(userIDStr)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return &domain.TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   domain.Role(roleStr),
	}, nil
}

// expireDuration คืนระยะเวลาหมดอายุของ token
func (s *tokenService) expireDuration() time.Duration {
	return time.Duration(s.expireHours) * time.Hour
}
