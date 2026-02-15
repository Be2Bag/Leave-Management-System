package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string // พอร์ตที่ server ทำงาน (default: 8080)

	MongoURI    string // MongoDB connection string
	MongoDBName string // ชื่อฐานข้อมูล

	JWTSecret      string // คีย์ลับสำหรับ sign JWT token (ต้องเปลี่ยนใน production!)
	JWTExpireHours string // จำนวนชั่วโมงก่อน token หมดอายุ

	CORSOrigins string // อนุญาต origins (default: * สำหรับ development เท่านั้น)
}

func Load() (*Config, error) {
	godotenv.Load() //nolint:errcheck // .env file is optional

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:    getEnv("MONGO_DB_NAME", "leave_management"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpireHours: getEnv("JWT_EXPIRE_HOURS", "24"),
		CORSOrigins:    getEnv("CORS_ORIGINS", "*"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET ต้องถูกกำหนดค่า (environment variable หรือ .env)")
	}

	const minSecretLength = 32
	if len(cfg.JWTSecret) < minSecretLength {
		return nil, fmt.Errorf("JWT_SECRET ต้องมีความยาวอย่างน้อย %d ตัวอักษร (ปัจจุบัน: %d)", minSecretLength, len(cfg.JWTSecret))
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
