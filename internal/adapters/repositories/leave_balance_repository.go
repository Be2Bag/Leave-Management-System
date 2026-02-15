package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github/be2bag/leave-management-system/internal/core/domain"
	"github/be2bag/leave-management-system/internal/core/ports"
	"github/be2bag/leave-management-system/internal/infrastructure/database"
)

type leaveBalanceRepository struct {
	collection *mongo.Collection
}

func NewLeaveBalanceRepository(db *database.MongoDB) ports.LeaveBalanceRepository {
	col := db.Database.Collection("leave_balances")

	// สร้าง compound unique index — ป้องกัน duplicate ยอดวันลา (user + type + year)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "leave_type", Value: 1},
			{Key: "year", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := col.Indexes().CreateOne(context.Background(), indexModel); err != nil {
		log.Printf("คำเตือน: สร้าง index leave_balances ไม่สำเร็จ: %v", err)
	}

	return &leaveBalanceRepository{collection: col}
}

// FindByUserID ค้นหายอดวันลาทั้งหมดของผู้ใช้
func (r *leaveBalanceRepository) FindByUserID(ctx context.Context, userID domain.ID) ([]domain.LeaveBalance, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ค้นหายอดวันลาล้มเหลว: %w", err)
	}

	var balances []domain.LeaveBalance
	if err := cursor.All(ctx, &balances); err != nil {
		return nil, fmt.Errorf("อ่านข้อมูลยอดวันลาล้มเหลว: %w", err)
	}

	return balances, nil
}

// ReservePending จองวันลาแบบ atomic ตอน submit ใบลา
func (r *leaveBalanceRepository) ReservePending(
	ctx context.Context,
	userID domain.ID,
	leaveType domain.LeaveType,
	year int,
	days float64,
) error {
	filter := bson.M{
		"user_id":    userID,
		"leave_type": leaveType,
		"year":       year,
		// Atomic condition: used + pending + requested <= total
		"$expr": bson.M{
			"$lte": bson.A{
				bson.M{"$add": bson.A{"$used_days", "$pending_days", days}},
				"$total_days",
			},
		},
	}

	update := bson.M{
		"$inc": bson.M{"pending_days": days},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("จองวันลาล้มเหลว: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrInsufficientBalance
	}

	return nil
}

// ConfirmPending ยืนยันวันลาแบบ atomic ตอนอนุมัติใบลา
func (r *leaveBalanceRepository) ConfirmPending(
	ctx context.Context,
	userID domain.ID,
	leaveType domain.LeaveType,
	year int,
	days float64,
) error {
	filter := bson.M{
		"user_id":    userID,
		"leave_type": leaveType,
		"year":       year,
	}

	update := bson.M{
		"$inc": bson.M{
			"used_days":    days,
			"pending_days": -days,
		},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ยืนยันวันลาล้มเหลว: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrLeaveBalanceNotFound
	}

	return nil
}

// ReleasePending ปล่อยวันลาที่จองไว้แบบ atomic ตอนปฏิเสธใบลา
func (r *leaveBalanceRepository) ReleasePending(
	ctx context.Context,
	userID domain.ID,
	leaveType domain.LeaveType,
	year int,
	days float64,
) error {
	filter := bson.M{
		"user_id":    userID,
		"leave_type": leaveType,
		"year":       year,
	}

	update := bson.M{
		"$inc": bson.M{"pending_days": -days},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ปล่อยวันลาที่จองไว้ล้มเหลว: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrLeaveBalanceNotFound
	}

	return nil
}
