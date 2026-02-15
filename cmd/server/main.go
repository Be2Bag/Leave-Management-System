package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	_ "github/be2bag/leave-management-system/docs"
	"github/be2bag/leave-management-system/internal/adapters/handlers"
	apphttp "github/be2bag/leave-management-system/internal/adapters/http"
	"github/be2bag/leave-management-system/internal/adapters/repositories"
	"github/be2bag/leave-management-system/internal/config"
	"github/be2bag/leave-management-system/internal/core/services"
	"github/be2bag/leave-management-system/internal/infrastructure/database"
	"github/be2bag/leave-management-system/pkg/validator"
)

// @title           Leave Management System API
// @version         1.0
// @description     ระบบจัดการการลาของพนักงาน — Backend API สำหรับยื่นใบลา อนุมัติ ปฏิเสธ และดูยอดวันลาคงเหลือ

// @contact.name   API Support
// @contact.email  support@company.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description กรุณาใส่ Bearer token เช่น "Bearer eyJhbGciOiJIUzI1NiIs..."

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("โหลด configuration ล้มเหลว: %w", err)
	}
	db, err := database.NewMongoDB(cfg)
	if err != nil {
		return fmt.Errorf("เชื่อมต่อ MongoDB ล้มเหลว: %w", err)
	}
	defer closeDB(db)
	log.Println("✅ เชื่อมต่อ MongoDB สำเร็จ")

	userRepo := repositories.NewUserRepository(db)
	balanceRepo := repositories.NewLeaveBalanceRepository(db)
	requestRepo := repositories.NewLeaveRequestRepository(db)

	jwtExpireHours := parseJWTExpireHours(cfg.JWTExpireHours)
	tokenService := services.NewTokenService(cfg.JWTSecret, jwtExpireHours)
	authService := services.NewAuthService(userRepo, tokenService)
	leaveService := services.NewLeaveService(requestRepo, balanceRepo)

	validate := validator.New()

	authHandler := handlers.NewAuthHandler(authService, validate)
	leaveHandler := handlers.NewLeaveHandler(leaveService, validate)

	app := createFiberApp(cfg.CORSOrigins)

	app.Get("/swagger/*", swagger.HandlerDefault)
	apphttp.SetupRouter(app, authHandler, leaveHandler, tokenService)

	go gracefulShutdown(app)

	log.Printf(" Swagger UI: http://localhost:%s/swagger/index.html", cfg.ServerPort)
	log.Printf("🚀 Leave Management System API กำลังทำงานที่พอร์ต %s", cfg.ServerPort)
	return app.Listen(":" + cfg.ServerPort)
}

func createFiberApp(corsOrigins string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:   "Leave Management System API",
		BodyLimit: 1 * 1024 * 1024, // จำกัดขนาด request body 1MB
	})

	app.Use(recover.New()) // จับ panic ป้องกัน server crash
	app.Use(logger.New())  // บันทึก HTTP request log
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	return app
}

func gracefulShutdown(app *fiber.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ กำลังปิด server...")
	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		log.Printf("ปิด server ไม่สำเร็จ: %v", err)
	}
}

func closeDB(db *database.MongoDB) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := db.Close(ctx); err != nil {
		log.Printf("ปิดการเชื่อมต่อ MongoDB ไม่สำเร็จ: %v", err)
	}
}

func parseJWTExpireHours(s string) int {
	hours, err := strconv.Atoi(s)
	if err != nil {
		return 24
	}
	return hours
}
