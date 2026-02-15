package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {

		c.Set("X-Content-Type-Options", "nosniff")                                // ป้องกัน MIME type sniffing
		c.Set("X-Frame-Options", "DENY")                                          // ป้องกัน Clickjacking
		c.Set("X-XSS-Protection", "1; mode=block")                                // ป้องกัน XSS (สำหรับ browser เก่า)
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains") // บังคับใช้ HTTPS (HSTS)
		c.Set("Cache-Control", "no-store")                                        // ป้องกันการ cache ข้อมูลสำคัญ
		if strings.HasPrefix(c.Path(), "/swagger") {                              // จำกัด resource ที่โหลดได้ — Swagger UI ต้องใช้ CSP ที่ผ่อนปรนกว่า
			c.Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline' https://unpkg.com; "+
					"img-src 'self' data: https://validator.swagger.io; "+
					"connect-src 'self'")
		} else {
			c.Set("Content-Security-Policy", "default-src 'none'")
		}

		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")             // ควบคุมข้อมูล referrer ที่ส่งออกไป
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()") // ปิดการเข้าถึง hardware APIs ที่ไม่จำเป็น

		return c.Next()
	}
}
