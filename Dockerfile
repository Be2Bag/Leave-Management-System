# ─── Stage 1: Build ──────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# คัดลอก dependency files ก่อน — ใช้ Docker layer cache
COPY go.mod go.sum ./
RUN go mod download

# คัดลอก source code ทั้งหมด
COPY . .

# คอมไพล์เป็น static binary (ไม่ต้องการ C libraries)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# ─── Stage 2: Runtime ───────────────────────────────────────────────────
FROM alpine:3.21

# ติดตั้ง CA certificates สำหรับ HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

# สร้าง non-root user เพื่อความปลอดภัย
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# คัดลอก binary จาก build stage
COPY --from=builder /server .

# ใช้ non-root user
USER appuser

# เปิดพอร์ต
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# รัน server
CMD ["./server"]
