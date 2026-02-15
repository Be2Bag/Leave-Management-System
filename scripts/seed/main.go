package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// ─── Seed Script ────────────────────────────────────────────────────────
// สร้างข้อมูลเริ่มต้นสำหรับทดสอบระบบ
// - 1 Manager: manager@company.com / password123
// - 1 Employee: employee@company.com / password123
// - ยอดวันลาเริ่มต้นสำหรับทั้งสองคน
//
// วิธีใช้: go run scripts/seed/main.go
// ─────────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load() //nolint:errcheck // .env file is optional

	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGO_DB_NAME", "leave_management")

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("เชื่อมต่อ MongoDB ล้มเหลว: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("ปิดการเชื่อมต่อล้มเหลว: %v", err)
		}
	}()

	ctx := context.Background()
	db := client.Database(dbName)

	fmt.Println("🌱 เริ่มสร้างข้อมูลเริ่มต้น...")

	dropCollections(ctx, db)

	managerID := uuid.New()
	employeeID := uuid.New()

	createUsers(ctx, db, managerID, employeeID)
	createLeaveBalances(ctx, db, managerID, employeeID)

	fmt.Println("")
	fmt.Println("✅ สร้างข้อมูลเริ่มต้นสำเร็จ!")
	fmt.Println("")
	fmt.Println("📋 ข้อมูลสำหรับทดสอบ:")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Println("👔 Manager:")
	fmt.Println("   Email:    manager@company.com")
	fmt.Println("   Password: password123")
	fmt.Printf("   UserID:   %s\n", managerID)
	fmt.Println("")
	fmt.Println("👤 Employee:")
	fmt.Println("   Email:    employee@company.com")
	fmt.Println("   Password: password123")
	fmt.Printf("   UserID:   %s\n", employeeID)
	fmt.Println("─────────────────────────────────────────────────")
}

// dropCollections ลบ collections ทั้งหมดเพื่อเริ่มต้นใหม่
func dropCollections(ctx context.Context, db *mongo.Database) {
	collections := []string{"users", "leave_balances", "leave_requests"}
	for _, name := range collections {
		if err := db.Collection(name).Drop(ctx); err != nil {
			log.Printf("คำเตือน: ลบ collection %s ไม่สำเร็จ: %v", name, err)
		}
	}
	fmt.Println("🗑️  ลบข้อมูลเก่าสำเร็จ")
}

// createUsers สร้างผู้ใช้ตัวอย่าง (Manager + Employee)
func createUsers(ctx context.Context, db *mongo.Database, managerID, employeeID uuid.UUID) {
	managerHash := hashPassword("password123")
	employeeHash := hashPassword("password123")
	now := time.Now()

	users := []interface{}{
		bson.M{
			"_id":           managerID,
			"first_name":    "สมชาย",
			"last_name":     "ผู้จัดการ",
			"full_name":     "สมชาย ผู้จัดการ",
			"email":         "manager@company.com",
			"password_hash": managerHash,
			"role":          "manager",
			"created_at":    now,
			"updated_at":    now,
		},
		bson.M{
			"_id":           employeeID,
			"first_name":    "สมหญิง",
			"last_name":     "พนักงาน",
			"full_name":     "สมหญิง พนักงาน",
			"email":         "employee@company.com",
			"password_hash": employeeHash,
			"role":          "employee",
			"created_at":    now,
			"updated_at":    now,
		},
	}

	col := db.Collection("users")
	if _, err := col.InsertMany(ctx, users); err != nil {
		log.Fatalf("สร้างผู้ใช้ล้มเหลว: %v", err)
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := col.Indexes().CreateOne(ctx, indexModel); err != nil {
		log.Printf("คำเตือน: สร้าง index email ไม่สำเร็จ: %v", err)
	}

	fmt.Println("👥 สร้างผู้ใช้ตัวอย่างสำเร็จ (Manager + Employee)")
}

// createLeaveBalances สร้างยอดวันลาเริ่มต้น
func createLeaveBalances(ctx context.Context, db *mongo.Database, managerID, employeeID uuid.UUID) {
	now := time.Now()
	year := now.Year()

	userIDs := []uuid.UUID{managerID, employeeID}
	leaveTypes := []struct {
		Type      string
		TotalDays float64
	}{
		{"sick_leave", 30},
		{"annual_leave", 15},
		{"personal_leave", 10},
	}

	var balances []interface{}
	for _, userID := range userIDs {
		for _, lt := range leaveTypes {
			balances = append(balances, bson.M{
				"_id":        uuid.New(),
				"user_id":    userID,
				"leave_type": lt.Type,
				"total_days": lt.TotalDays,
				"used_days":  0,
				"year":       year,
				"created_at": now,
				"updated_at": now,
			})
		}
	}

	col := db.Collection("leave_balances")
	if _, err := col.InsertMany(ctx, balances); err != nil {
		log.Fatalf("สร้างยอดวันลาล้มเหลว: %v", err)
	}

	// สร้าง compound unique index
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "leave_type", Value: 1},
			{Key: "year", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := col.Indexes().CreateOne(ctx, indexModel); err != nil {
		log.Printf("คำเตือน: สร้าง index leave_balances ไม่สำเร็จ: %v", err)
	}

	fmt.Println("📊 สร้างยอดวันลาเริ่มต้นสำเร็จ")
}

// hashPassword เข้ารหัสรหัสผ่านด้วย bcrypt
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("เข้ารหัสรหัสผ่านล้มเหลว: %v", err)
	}
	return string(hash)
}

// getEnv อ่านค่า environment variable
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
