package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github/be2bag/leave-management-system/internal/config"
)

const connectTimeout = 10 * time.Second

type MongoDB struct {
	Client   *mongo.Client   // client สำหรับจัดการ connection
	Database *mongo.Database // database instance สำหรับ CRUD operations
}

// NewMongoDB สร้าง MongoDB connection ใหม่
func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	clientOpts := options.Client().ApplyURI(cfg.MongoURI)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("เชื่อมต่อ MongoDB ล้มเหลว: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping MongoDB ล้มเหลว: %w", err)
	}

	db := client.Database(cfg.MongoDBName)
	return &MongoDB{Client: client, Database: db}, nil
}

// Close ปิดการเชื่อมต่อ MongoDB อย่างปลอดภัย
func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("ปิดการเชื่อมต่อ MongoDB ล้มเหลว: %w", err)
	}
	return nil
}
