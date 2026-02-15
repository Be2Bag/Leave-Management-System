package domain

import "github.com/google/uuid"

// ID ใช้เป็น primary key ของทุก entity
type ID = uuid.UUID

// NewID สร้าง ID ใหม่แบบ UUID v4
func NewID() ID {
	return uuid.New()
}

// ParseID แปลงข้อความเป็น ID
func ParseID(s string) (ID, error) {
	return uuid.Parse(s)
}
