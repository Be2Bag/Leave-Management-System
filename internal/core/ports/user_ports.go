package ports

import (
	"context"

	"github/be2bag/leave-management-system/internal/core/domain"
)

type UserRepository interface {
	// FindByEmail ค้นหาผู้ใช้จากอีเมล
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
