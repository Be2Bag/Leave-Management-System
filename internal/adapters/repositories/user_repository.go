package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
	"github/be2bag/leave-management-system/internal/infrastructure/database"
)

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) ports.UserRepository {
	col := db.Database.Collection("users")

	// สร้าง unique index สำหรับ email — ป้องกัน duplicate email ในระดับ database
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := col.Indexes().CreateOne(context.Background(), indexModel); err != nil {
		log.Printf("คำเตือน: สร้าง index email ไม่สำเร็จ: %v", err)
	}

	return &userRepository{collection: col}
}

// FindByEmail ค้นหาผู้ใช้จากอีเมล
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	filter := bson.M{"email": email}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("ค้นหาผู้ใช้จากอีเมลล้มเหลว: %w", err)
	}

	return &user, nil
}
