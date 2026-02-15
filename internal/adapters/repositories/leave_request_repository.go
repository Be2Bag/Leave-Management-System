package repositories

import (
	"context"
	"errors"
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

type leaveRequestRepository struct {
	collection *mongo.Collection
}

func NewLeaveRequestRepository(db *database.MongoDB) ports.LeaveRequestRepository {
	col := db.Database.Collection("leave_requests")
	createLeaveRequestIndexes(col)
	return &leaveRequestRepository{collection: col}
}

// createLeaveRequestIndexes สร้าง indexes สำหรับ collection leave_requests
func createLeaveRequestIndexes(col *mongo.Collection) {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},                                                             // ค้นหาตาม user
		{Keys: bson.D{{Key: "status", Value: 1}}},                                                              // ค้นหาตามสถานะ
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "start_date", Value: 1}, {Key: "end_date", Value: 1}}}, // overlap check
	}

	for _, idx := range indexes {
		if _, err := col.Indexes().CreateOne(context.Background(), idx); err != nil {
			log.Printf("คำเตือน: สร้าง index leave_requests ไม่สำเร็จ: %v", err)
		}
	}
}

// Create สร้างคำขอลาใหม่
func (r *leaveRequestRepository) Create(ctx context.Context, request *domain.LeaveRequest) error {
	_, err := r.collection.InsertOne(ctx, request)
	if err != nil {
		return fmt.Errorf("สร้างคำขอลาล้มเหลว: %w", err)
	}
	return nil
}

// FindByID ค้นหาคำขอลาจากรหัส
func (r *leaveRequestRepository) FindByID(ctx context.Context, id domain.ID) (*domain.LeaveRequest, error) {
	var request domain.LeaveRequest
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrRequestNotFound
		}
		return nil, fmt.Errorf("ค้นหาคำขอลาล้มเหลว: %w", err)
	}

	return &request, nil
}

// FindByUserID ค้นหาคำขอลาทั้งหมดของผู้ใช้ (เรียงจากใหม่สุด, รองรับ pagination)
func (r *leaveRequestRepository) FindByUserID(
	ctx context.Context,
	userID domain.ID,
	params domain.PaginationParams,
) (*domain.PaginatedResult[domain.LeaveRequest], error) {
	filter := bson.M{"user_id": userID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("นับจำนวนคำขอลาล้มเหลว: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(params.Offset()).
		SetLimit(params.Limit())

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("ค้นหาคำขอลาตาม user ล้มเหลว: %w", err)
	}

	var requests []domain.LeaveRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, fmt.Errorf("อ่านข้อมูลคำขอลาล้มเหลว: %w", err)
	}

	return domain.NewPaginatedResult(requests, total, params), nil
}

// FindByStatus ค้นหาคำขอลาตามสถานะ (เรียงจากเก่าสุดก่อน สำหรับ FIFO processing, รองรับ pagination)
func (r *leaveRequestRepository) FindByStatus(
	ctx context.Context,
	status domain.LeaveStatus,
	params domain.PaginationParams,
) (*domain.PaginatedResult[domain.LeaveRequest], error) {
	filter := bson.M{"status": status}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("นับจำนวนคำขอลาล้มเหลว: %w", err)
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetSkip(params.Offset()).
		SetLimit(params.Limit())

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("ค้นหาคำขอลาตามสถานะล้มเหลว: %w", err)
	}

	var requests []domain.LeaveRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, fmt.Errorf("อ่านข้อมูลคำขอลาล้มเหลว: %w", err)
	}

	return domain.NewPaginatedResult(requests, total, params), nil
}

// Update อัปเดตคำขอลา (ใช้ ReplaceOne เพื่อแทนที่ทั้ง document)
func (r *leaveRequestRepository) Update(ctx context.Context, request *domain.LeaveRequest) error {
	filter := bson.M{"_id": request.ID}

	result, err := r.collection.ReplaceOne(ctx, filter, request)
	if err != nil {
		return fmt.Errorf("อัปเดตคำขอลาล้มเหลว: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrRequestNotFound
	}

	return nil
}

// UpdateWithStatusCheck อัปเดตคำขอลาแบบ atomic — อัปเดตเฉพาะเมื่อสถานะตรงกับที่คาดหวัง
func (r *leaveRequestRepository) UpdateWithStatusCheck(
	ctx context.Context,
	request *domain.LeaveRequest,
	expectedStatus domain.LeaveStatus,
) error {
	filter := bson.M{
		"_id":    request.ID,
		"status": expectedStatus,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, request)
	if err != nil {
		return fmt.Errorf("อัปเดตคำขอลาล้มเหลว: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrRequestAlreadyProcessed
	}

	return nil
}

// HasOverlap ตรวจสอบว่ามีคำขอลาซ้ำซ้อนกับช่วงวันที่ที่ระบุหรือไม่
func (r *leaveRequestRepository) HasOverlap(
	ctx context.Context,
	userID domain.ID,
	startDate, endDate time.Time,
	excludeID *domain.ID,
) (bool, error) {
	filter := bson.M{
		"user_id":    userID,
		"status":     bson.M{"$in": []string{string(domain.LeaveStatusPending), string(domain.LeaveStatusApproved)}},
		"start_date": bson.M{"$lte": endDate},   // ใบลาเริ่มก่อนหรือตรงกับวันสิ้นสุดที่ขอ
		"end_date":   bson.M{"$gte": startDate}, // ใบลาสิ้นสุดหลังหรือตรงกับวันเริ่มต้นที่ขอ
	}

	// กรณีแก้ไขใบลา — ไม่นับใบลาที่กำลังแก้ไขเอง
	if excludeID != nil {
		filter["_id"] = bson.M{"$ne": *excludeID}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("ตรวจสอบวันลาซ้ำซ้อนล้มเหลว: %w", err)
	}

	return count > 0, nil
}
