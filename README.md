# Leave Management System API

ระบบจัดการการลาของพนักงาน — Backend API สำหรับยื่นใบลา, อนุมัติ/ปฏิเสธ, และติดตามยอดวันลาคงเหลือ

## 📋 สารบัญ

- [เทคโนโลยีที่ใช้](#-เทคโนโลยีที่ใช้)
- [สถาปัตยกรรม](#-สถาปัตยกรรม)
- [โครงสร้างโปรเจค](#-โครงสร้างโปรเจค)
- [วิธีติดตั้งและรัน](#-วิธีติดตั้งและรัน)
- [ข้อมูลทดสอบ (Seed Data)](#-ข้อมูลทดสอบ-seed-data)
- [API Endpoints](#-api-endpoints)
- [Business Assumptions](#-business-assumptions)
- [Overlap Rule](#-overlap-rule)
- [ER Diagram](#-er-diagram)
- [Data Dictionary](#-data-dictionary)
- [Indexes & Query Examples](#-indexes--query-examples)
- [เหตุผลในการออกแบบ](#-เหตุผลในการออกแบบ)
- [ความปลอดภัย](#-ความปลอดภัย)
- [การทดสอบและคุณภาพโค้ด](#-การทดสอบและคุณภาพโค้ด)
- [ข้อจำกัดที่ทราบ](#-ข้อจำกัดที่ทราบ)

---

## 🛠 เทคโนโลยีที่ใช้

| เทคโนโลยี | เวอร์ชัน | หน้าที่ |
|---|---|---|
| **Go** | 1.25.6 | ภาษาหลัก |
| **Fiber v2** | 2.52.11 | HTTP Framework (เร็ว, คล้าย Express.js) |
| **MongoDB** | 7.x | ฐานข้อมูล NoSQL |
| **mongo-driver v2** | 2.5.0 | MongoDB Go Driver |
| **JWT** | v5 | ยืนยันตัวตน (golang-jwt) |
| **bcrypt** | x/crypto | เข้ารหัสรหัสผ่าน |
| **validator v10** | 10.30.1 | ตรวจสอบข้อมูลขาเข้า |
| **Swagger** | - | เอกสาร API อัตโนมัติ |
| **golangci-lint** | 1.64.8 | ตรวจสอบคุณภาพโค้ดและบังคับ Architecture Rules |
| **Docker** | - | Containerization |

---

## 🏗 สถาปัตยกรรม

### Hexagonal Architecture (Ports & Adapters)

```
                    ┌─────────────────────────────────────────┐
                    │              cmd/server/                 │
                    │        (จุดเริ่มต้นของแอป)                │
                    └──────────┬──────────┬───────────────────┘
                               │          │
            ┌──────────────────▼──┐  ┌────▼──────────────────┐
            │   Primary Adapters  │  │  Secondary Adapters    │
            │  ┌────────────────┐ │  │ ┌────────────────────┐ │
            │  │   Handlers     │ │  │ │   Repositories     │ │
            │  │  (HTTP → Svc)  │ │  │ │  (Svc → MongoDB)   │ │
            │  └───────┬────────┘ │  │ └────────┬───────────┘ │
            │          │          │  │           │             │
            └──────────┼──────────┘  └───────────┼─────────────┘
                       │                         │
            ┌──────────▼─────────────────────────▼─────────────┐
            │                    PORTS                          │
            │           (Interface / สัญญาที่ตกลงกัน)            │
            ├───────────────────────────────────────────────────┤
            │                  SERVICES                         │
            │           (Business Logic / Use Cases)            │
            ├───────────────────────────────────────────────────┤
            │                   DOMAIN                          │
            │     (Entities, Value Objects, กฎทางธุรกิจ)         │
            └───────────────────────────────────────────────────┘
```

**กฎการพึ่งพา (Dependency Rule):** ลูกศรชี้เข้าด้านในเสมอ — layer ภายนอกพึ่งพา layer ภายใน ไม่เคยกลับกัน

### รูปแบบการออกแบบ (Design Patterns)

| รูปแบบ | คำอธิบาย |
|---|---|
| **Hexagonal Architecture** | แยก business logic ออกจาก infrastructure ทำให้เปลี่ยน framework/database ได้โดยไม่กระทบ logic |
| **Dependency Injection** | inject dependencies ผ่าน constructor ทำให้ทดสอบง่ายด้วย mock |
| **Repository Pattern** | abstraction สำหรับ data access — service ไม่รู้จัก database โดยตรง |
| **Domain-Driven Design (Lite)** | entities, value objects, domain errors อยู่ใน layer ในสุด |
| **DTO Pattern** | แยก API contract (request/response) ออกจาก domain model |

---

## 📁 โครงสร้างโปรเจค

```
leave-management-system/
├── cmd/server/main.go                 # จุดเริ่มต้น — ประกอบ dependencies ทั้งหมด
├── internal/
│   ├── core/                          # ── Business Logic (ไม่รู้จัก framework) ──
│   │   ├── domain/                    # Entities, Enums, กฎทางธุรกิจ, Errors
│   │   │   ├── id.go                  # UUID type alias
│   │   │   ├── role.go                # บทบาทผู้ใช้ (employee/manager)
│   │   │   ├── leave_status.go        # สถานะใบลา (pending/approved/rejected)
│   │   │   ├── leave_type.go          # ประเภทการลา (ป่วย/พักร้อน/กิจส่วนตัว)
│   │   │   ├── user.go                # Entity ผู้ใช้
│   │   │   ├── leave_balance.go       # Entity ยอดวันลา
│   │   │   ├── leave_request.go       # Entity ใบลา
│   │   │   ├── pagination.go          # โครงสร้างข้อมูลสำหรับแบ่งหน้า
│   │   │   ├── token_claims.go        # โครงสร้างข้อมูล JWT Claims
│   │   │   ├── errors.go              # Domain errors ทั้งหมด
│   │   │   └── domain_test.go         # ทดสอบ domain logic
│   │   ├── ports/                     # Interfaces / สัญญาระหว่าง layer
│   │   │   ├── auth_ports.go          # Interface สำหรับ Auth (Login)
│   │   │   ├── leave_ports.go         # Interface สำหรับจัดการลาและ Repositories
│   │   │   └── user_ports.go          # Interface สำหรับจัดการผู้ใช้
│   │   └── services/                  # ตัวดำเนินการ Business Logic
│   │       ├── auth_service.go        # เข้าสู่ระบบ
│   │       ├── token_service.go       # สร้างและตรวจสอบ JWT
│   │       ├── leave_service.go       # ยื่น/อนุมัติ/ปฏิเสธใบลา
│   │       ├── auth_service_test.go   # ทดสอบ auth service
│   │       ├── leave_service_test.go  # ทดสอบ leave service
│   │       └── mocks_test.go          # Mock repositories สำหรับทดสอบ
│   ├── adapters/                      # ── ตัวเชื่อมต่อกับโลกภายนอก ──
│   │   ├── dto/                       # โครงสร้างข้อมูลสำหรับ API (request/response)
│   │   │   ├── auth_dto.go            # DTO สำหรับ Login
│   │   │   ├── leave_dto.go           # DTO สำหรับจัดการลา
│   │   │   └── response.go            # รูปแบบ response มาตรฐาน
│   │   ├── handlers/                  # HTTP Handlers (รับ request → เรียก service)
│   │   │   ├── auth_handler.go        # จัดการ endpoint ยืนยันตัวตน
│   │   │   ├── leave_handler.go       # จัดการ endpoint การลา
│   │   │   └── error_handler.go       # แปลง domain error → HTTP response
│   │   ├── http/                      # Router และ Middleware
│   │   │   ├── router.go              # กำหนดเส้นทาง API ทั้งหมด
│   │   │   └── middleware/
│   │   │       ├── auth.go            # ตรวจสอบ JWT token และสิทธิ์ตาม role
│   │   │       └── security.go        # Security headers (XSS, CSRF ฯลฯ)
│   │   └── repositories/             # เชื่อมต่อกับ MongoDB
│   │       ├── user_repository.go     # อ่านข้อมูลผู้ใช้
│   │       ├── leave_balance_repository.go  # จัดการยอดวันลา (atomic operations)
│   │       └── leave_request_repository.go  # จัดการใบลา
│   ├── config/
│   │   └── config.go                  # โหลด environment variables
│   └── infrastructure/database/
│       └── mongodb.go                 # เชื่อมต่อ MongoDB
├── pkg/validator/
│   └── validator.go                   # ตัวตรวจสอบข้อมูลขาเข้า (ใช้ร่วมกันทั้งโปรเจค)
├── scripts/seed/
│   └── main.go                        # สร้างข้อมูลทดสอบ (Manager + Employee)
├── docs/                              # Swagger API Docs (สร้างอัตโนมัติ)
├── .golangci.yml                      # กฎ lint บังคับ Hexagonal Architecture
├── Dockerfile                         # Multi-stage Docker build
├── docker-compose.yml                 # รัน App + MongoDB พร้อมกัน
└── .env.example                       # ตัวอย่างไฟล์ environment variables
```

---

## 🚀 วิธีติดตั้งและรัน

### สิ่งที่ต้องมี

- **Go** 1.25.6+
- **MongoDB** 7.x (ติดตั้งเอง หรือใช้ Docker)
- **golangci-lint** (ไม่บังคับ — สำหรับตรวจสอบคุณภาพโค้ด)

### วิธีที่ 1: Docker Compose (แนะนำ)

```bash
# Clone โปรเจค
git clone <repository-url>
cd leave-management-system

# รัน App + MongoDB ด้วย Docker Compose
docker-compose up -d

# ดู logs
docker-compose logs -f app

# สร้างข้อมูลทดสอบ (ในอีก terminal)
docker-compose exec app sh -c "cd /app && go run scripts/seed/main.go"
```

### วิธีที่ 2: รันบนเครื่อง (Local Development)

```bash
# 1. Clone โปรเจค
git clone <repository-url>
cd leave-management-system

# 2. ติดตั้ง dependencies
go mod download

# 3. สร้างไฟล์ .env จากตัวอย่าง
cp .env.example .env
# แก้ไขค่าใน .env ตามต้องการ (อย่าลืมเปลี่ยน JWT_SECRET)

# 4. ตรวจสอบว่า MongoDB กำลังทำงาน
# ถ้ายังไม่มี MongoDB ให้รันผ่าน Docker:
docker run -d --name mongodb -p 27017:27017 mongo:7

# 5. สร้างข้อมูลทดสอบ
go run scripts/seed/main.go

# 6. รัน server
go run cmd/server/main.go
```

Server จะทำงานที่ `http://localhost:8080`
Swagger UI จะอยู่ที่ `http://localhost:8080/swagger/`

---

## 🌱 ข้อมูลทดสอบ (Seed Data)

รัน seed script เพื่อสร้างผู้ใช้และยอดวันลาเริ่มต้น:

```bash
go run scripts/seed/main.go
```

จะสร้างข้อมูลดังนี้:

| บทบาท | อีเมล | รหัสผ่าน | ยอดวันลา |
|------|-------|---------|---------|
| Manager | manager@company.com | password123 | ลาป่วย 30 วัน, ลาพักร้อน 15 วัน, ลากิจ 10 วัน |
| Employee | employee@company.com | password123 | ลาป่วย 30 วัน, ลาพักร้อน 15 วัน, ลากิจ 10 วัน |

> 💡 รหัสผ่านถูก hash ด้วย bcrypt (cost 12) — ไม่ได้เก็บเป็น plain text

---

## 🔌 API Endpoints

### ยืนยันตัวตน (Public — ไม่ต้อง Login)

| Method | Endpoint | คำอธิบาย |
|--------|----------|---------|
| `POST` | `/api/v1/auth/login` | เข้าสู่ระบบ — รับ JWT token |

### จัดการการลา (ต้อง Login — Employee, Manager)

| Method | Endpoint | คำอธิบาย |
|--------|----------|---------|
| `POST` | `/api/v1/leaves/` | ยื่นใบลา |
| `GET` | `/api/v1/leaves/my-requests` | ดูประวัติใบลาของตนเอง (รองรับแบ่งหน้า) |
| `GET` | `/api/v1/leaves/my-balance` | ดูยอดวันลาคงเหลือ |

### สำหรับผู้จัดการ (ต้องเป็น Manager เท่านั้น)

| Method | Endpoint | คำอธิบาย |
|--------|----------|---------|
| `GET` | `/api/v1/manager/pending-requests` | ดูใบลารอการอนุมัติ (รองรับแบ่งหน้า) |
| `POST` | `/api/v1/manager/requests/:id/approve` | อนุมัติใบลา |
| `POST` | `/api/v1/manager/requests/:id/reject` | ปฏิเสธใบลา |

### อื่นๆ

| Method | Endpoint | คำอธิบาย |
|--------|----------|---------|
| `GET` | `/health` | ตรวจสอบสถานะ API server |
| `GET` | `/swagger/*` | เอกสาร API แบบ Swagger UI |

### ตัวอย่างการใช้งาน

<details>
<summary>🔑 เข้าสู่ระบบ (Login)</summary>

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "employee@company.com",
    "password": "password123"
  }'
```
</details>

<details>
<summary>📋 ยื่นใบลา</summary>

```bash
curl -X POST http://localhost:8080/api/v1/leaves/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "leave_type": "annual_leave",
    "start_date": "2026-03-01",
    "end_date": "2026-03-03",
    "reason": "ลาพักร้อนไปเที่ยวกับครอบครัว"
  }'
```
</details>

<details>
<summary>✅ อนุมัติใบลา (Manager)</summary>

```bash
curl -X POST http://localhost:8080/api/v1/manager/requests/<request-id>/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <manager-jwt-token>" \
  -d '{
    "note": "อนุมัติ"
  }'
```
</details>

<details>
<summary>❌ ปฏิเสธใบลา (Manager)</summary>

```bash
curl -X POST http://localhost:8080/api/v1/manager/requests/<request-id>/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <manager-jwt-token>" \
  -d '{
    "note": "กรุณาเลื่อนวันลา เนื่องจากตรงกับช่วงปิดงบ"
  }'
```
</details>

---

## 📌 Business Assumptions

สมมติฐานทางธุรกิจที่ระบบใช้ในการคำนวณและจัดการวันลา:

| หัวข้อ | พฤติกรรม | รายละเอียด |
|---|---|---|
| **การนับวันลา** | Inclusive (start–end) | นับรวมทั้งวันเริ่มต้นและวันสิ้นสุด เช่น 1–3 มี.ค. = **3 วัน** (`end - start + 1`) |
| **วันหยุดสุดสัปดาห์** | ไม่ตัดออก | นับเป็น **calendar days** — เสาร์-อาทิตย์ถูกนับเป็นวันลาด้วย ไม่มี logic แยกวันทำการ (business days) |
| **วันหยุดนักขัตฤกษ์** | ไม่ตัดออก | ระบบไม่มี holiday list — วันหยุดราชการถูกนับรวมเป็นวันลา |
| **Timezone** | UTC | วันที่ทั้งหมดถูกตีความเป็น UTC — `CalculateLeaveDays` normalize เป็น `time.UTC` และ `time.Parse("2006-01-02", ...)` ได้ผลลัพธ์เป็น UTC โดย default |
| **Half-day** | ยังไม่รองรับ | field `TotalDays` เป็น `float64` (โครงสร้างพร้อมรับทศนิยม) แต่ logic คำนวณได้เฉพาะจำนวน**เต็มวัน** — ไม่มี field `HalfDay` หรือ `Period` (เช้า/บ่าย) ใน request |

### ตัวอย่างการคำนวณ

```go
// CalculateLeaveDays — internal/core/domain/leave_request.go
func CalculateLeaveDays(startDate, endDate time.Time) float64 {
    start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
    end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
    days := end.Sub(start).Hours() / 24
    return days + 1 // +1 เพราะนับรวมวันเริ่มต้นด้วย (inclusive)
}
```

| ตัวอย่าง | start | end | ผลลัพธ์ |
|---|---|---|---|
| ลา 1 วัน | 2026-03-01 | 2026-03-01 | **1 วัน** |
| ลา 3 วัน | 2026-03-01 | 2026-03-03 | **3 วัน** |
| ลาข้ามสัปดาห์ | 2026-01-01 (พฤ) | 2026-01-05 (จ) | **5 วัน** (รวม ส-อา) |

---

## 🔀 Overlap Rule

ระบบตรวจสอบ **วันลาซ้ำซ้อน (overlap)** ก่อนอนุญาตให้ยื่นใบลาใหม่ — พนักงานคนเดียวกันจะยื่นใบลาที่มีช่วงวันซ้อนทับกันไม่ได้

### เงื่อนไขที่ถือว่าซ้อนทับ

ใบลาใหม่ `[new_start, new_end]` จะ **overlap** กับใบลาเดิม `[exist_start, exist_end]` เมื่อ:

```
exist_start <= new_end   AND   exist_end >= new_start
```

นี่คือ **standard interval overlap formula** — ครอบคลุมทุกกรณี:

| กรณี | ไทม์ไลน์ | ผลลัพธ์ |
|---|---|---|
| ซ้อนทับบางส่วน (ขวา) | `exist: [1-5]` vs `new: [3-8]` | ❌ Overlap |
| ซ้อนทับบางส่วน (ซ้าย) | `exist: [5-10]` vs `new: [3-7]` | ❌ Overlap |
| ใบลาใหม่ครอบ | `exist: [3-5]` vs `new: [1-10]` | ❌ Overlap |
| ใบลาเดิมครอบ | `exist: [1-10]` vs `new: [3-5]` | ❌ Overlap |
| วันเดียวกัน | `exist: [5-5]` vs `new: [5-5]` | ❌ Overlap |
| ติดกันพอดี (ไม่ซ้อน) | `exist: [1-5]` vs `new: [6-10]` | ✅ OK |
| ไม่เกี่ยวกัน | `exist: [1-3]` vs `new: [7-10]` | ✅ OK |

### สถานะที่ตรวจสอบ

ตรวจเฉพาะใบลาที่มีสถานะ **`pending`** หรือ **`approved`** เท่านั้น — ใบลาที่ถูก `rejected` แล้วจะไม่นับ

### MongoDB Query ที่ใช้

```javascript
// HasOverlap — ตรวจสอบวันลาซ้ำซ้อน
db.leave_requests.countDocuments({
  user_id:    <userID>,
  status:     { $in: ["pending", "approved"] },
  start_date: { $lte: <new_end> },    // ใบลาเดิมเริ่มก่อนวันสิ้นสุดใหม่
  end_date:   { $gte: <new_start> }   // ใบลาเดิมจบหลังวันเริ่มต้นใหม่
})
// ถ้า count > 0 → ซ้ำซ้อน → reject ทันที
```

### ตัวอย่าง

```
ใบลาเดิม (approved):     2026-03-03  ───  2026-03-07

ยื่นใหม่  2026-03-01 ~ 03-02   ✅ OK (จบก่อนใบเดิมเริ่ม)
ยื่นใหม่  2026-03-01 ~ 03-03   ❌ Overlap (ชนวันเริ่มต้นพอดี)
ยื่นใหม่  2026-03-05 ~ 03-10   ❌ Overlap (ซ้อนกลาง)
ยื่นใหม่  2026-03-08 ~ 03-10   ✅ OK (เริ่มหลังใบเดิมจบ)
```

---

## 📊 ER Diagram

```
┌─────────────────────────┐       ┌──────────────────────────┐
│         Users            │       │     Leave Balances        │
│       (ผู้ใช้งาน)         │       │     (ยอดวันลา)            │
├─────────────────────────┤       ├──────────────────────────┤
│ _id: UUID (PK)          │──┐    │ _id: UUID (PK)           │
│ first_name: string      │  │    │ user_id: UUID (FK)       │──┐
│ last_name: string       │  ├───▶│ leave_type: string       │  │
│ full_name: string       │  │    │ total_days: float64      │  │
│ email: string (unique)  │  │    │ used_days: float64       │  │
│ password_hash: string   │  │    │ pending_days: float64    │  │
│ role: string             │  │    │ year: int                │  │
│ created_at: datetime    │  │    │ created_at: datetime     │  │
│ updated_at: datetime    │  │    │ updated_at: datetime     │  │
└─────────────────────────┘  │    └──────────────────────────┘  │
                              │                                   │
                              │    ┌──────────────────────────┐  │
                              │    │    Leave Requests         │  │
                              │    │    (ใบลา)                 │  │
                              │    ├──────────────────────────┤  │
                              │    │ _id: UUID (PK)           │  │
                              └───▶│ user_id: UUID (FK)       │──┘
                                   │ leave_type: string       │
                                   │ start_date: datetime     │
                                   │ end_date: datetime       │
                                   │ total_days: float64      │
                                   │ reason: string           │
                                   │ status: string           │
                                   │ reviewer_id: UUID (FK)   │───▶ Users._id
                                   │ review_note: string      │
                                   │ reviewed_at: datetime    │
                                   │ created_at: datetime     │
                                   │ updated_at: datetime     │
                                   └──────────────────────────┘
```

> 📎 รายละเอียด fields แต่ละ collection ดูที่ [Data Dictionary](#-data-dictionary) / รายละเอียด indexes ดูที่ [Indexes & Query Examples](#-indexes--query-examples)

---

## 📖 Data Dictionary

รายละเอียด fields ทั้งหมดของแต่ละ collection ใน MongoDB

### Collection: `users`

| Field | BSON Key | Type | Constraint | คำอธิบาย |
|---|---|---|---|---|
| รหัสผู้ใช้ | `_id` | `UUID` | **PK** | UUID v4 สร้างฝั่ง application |
| ชื่อจริง | `first_name` | `string` | required | ชื่อจริงของพนักงาน |
| นามสกุล | `last_name` | `string` | required | นามสกุลของพนักงาน |
| ชื่อเต็ม | `full_name` | `string` | auto | `first_name + " " + last_name` สร้างอัตโนมัติ |
| อีเมล | `email` | `string` | **unique**, required | ใช้เป็น username สำหรับ Login |
| รหัสผ่าน (hash) | `password_hash` | `string` | required | bcrypt hash (cost 12) — ไม่ส่งกลับใน JSON |
| บทบาท | `role` | `string` | required | `"employee"` \| `"manager"` |
| วันที่สร้าง | `created_at` | `datetime` | auto | |
| วันที่แก้ไขล่าสุด | `updated_at` | `datetime` | auto | |

### Collection: `leave_balances`

| Field | BSON Key | Type | Constraint | คำอธิบาย |
|---|---|---|---|---|
| รหัสยอดวันลา | `_id` | `UUID` | **PK** | |
| รหัสพนักงาน | `user_id` | `UUID` | **FK → users** | เจ้าของยอดวันลา |
| ประเภทการลา | `leave_type` | `string` | required | `"sick_leave"` \| `"annual_leave"` \| `"personal_leave"` |
| วันลาทั้งหมด | `total_days` | `float64` | required | โควตาวันลาต่อปี (เช่น ป่วย 30, พักร้อน 15, กิจ 10) |
| วันลาที่ใช้แล้ว | `used_days` | `float64` | default: 0 | จำนวนวันที่อนุมัติแล้ว |
| วันลาที่จองไว้ | `pending_days` | `float64` | default: 0 | จำนวนวันที่รออนุมัติ (Reserve → Confirm/Release) |
| ปี | `year` | `int` | required | ปี พ.ศ./ค.ศ. ที่ยอดนี้ใช้ได้ |
| วันที่สร้าง | `created_at` | `datetime` | auto | |
| วันที่แก้ไขล่าสุด | `updated_at` | `datetime` | auto | |

> **Compound Unique:** `(user_id, leave_type, year)` — 1 คน + 1 ประเภท + 1 ปี = 1 document เท่านั้น

### Collection: `leave_requests`

| Field | BSON Key | Type | Constraint | คำอธิบาย |
|---|---|---|---|---|
| รหัสใบลา | `_id` | `UUID` | **PK** | |
| รหัสพนักงาน | `user_id` | `UUID` | **FK → users** | ผู้ยื่นใบลา |
| ประเภทการลา | `leave_type` | `string` | required | `"sick_leave"` \| `"annual_leave"` \| `"personal_leave"` |
| วันเริ่มต้นลา | `start_date` | `datetime` | required | normalize เป็น UTC 00:00:00 |
| วันสิ้นสุดลา | `end_date` | `datetime` | required | normalize เป็น UTC 00:00:00 |
| จำนวนวันลา | `total_days` | `float64` | auto | คำนวณจาก `CalculateLeaveDays(start, end)` — inclusive |
| เหตุผลการลา | `reason` | `string` | required, 5-500 chars | |
| สถานะ | `status` | `string` | required | `"pending"` \| `"approved"` \| `"rejected"` |
| รหัสผู้อนุมัติ | `reviewer_id` | `UUID` | nullable, **FK → users** | Manager ที่ approve/reject — `null` ขณะ pending |
| หมายเหตุผู้อนุมัติ | `review_note` | `string` | optional | |
| วันที่อนุมัติ/ปฏิเสธ | `reviewed_at` | `datetime` | nullable | `null` ขณะ pending |
| วันที่ยื่นใบลา | `created_at` | `datetime` | auto | |
| วันที่แก้ไขล่าสุด | `updated_at` | `datetime` | auto | |

### Enum Values

| Enum | ค่าที่เป็นไปได้ | คำอธิบาย |
|---|---|---|
| **Role** | `employee`, `manager` | พนักงานยื่นลา / ผู้จัดการอนุมัติ-ปฏิเสธ |
| **LeaveType** | `sick_leave`, `annual_leave`, `personal_leave` | ลาป่วย (30 วัน), ลาพักร้อน (15 วัน), ลากิจ (10 วัน) |
| **LeaveStatus** | `pending`, `approved`, `rejected` | รออนุมัติ → อนุมัติ/ปฏิเสธ |

---

## 📇 Indexes & Query Examples

รายละเอียด MongoDB indexes ทั้งหมดที่ระบบสร้างอัตโนมัติเมื่อ application เริ่มทำงาน

### Collection: `users`

| Index | Fields | Type | วัตถุประสงค์ |
|---|---|---|---|
| `email_1` | `{ email: 1 }` | **Unique** | ป้องกันอีเมลซ้ำ + ใช้ค้นหาตอน Login |

```javascript
// Login — ค้นหาผู้ใช้จากอีเมล (ใช้ unique index)
db.users.findOne({ email: "somchai@company.com" })
```

### Collection: `leave_balances`

| Index | Fields | Type | วัตถุประสงค์ |
|---|---|---|---|
| `user_id_1_leave_type_1_year_1` | `{ user_id: 1, leave_type: 1, year: 1 }` | **Compound Unique** | ป้องกันยอดวันลาซ้ำ (1 user + 1 type + 1 year = 1 document) |

```javascript
// ดูยอดวันลาทั้งหมดของ user
db.leave_balances.find({ user_id: <userID> })

// ReservePending — จองวันลาแบบ atomic (ตรวจสอบยอดพอหรือไม่ในคำสั่งเดียว)
db.leave_balances.updateOne(
  {
    user_id: <userID>, leave_type: "annual_leave", year: 2026,
    $expr: { $lte: [{ $add: ["$used_days", "$pending_days", 3] }, "$total_days"] }
  },
  {
    $inc: { pending_days: 3 },
    $set: { updated_at: new Date() }
  }
)

// ConfirmPending — อนุมัติ: ย้าย pending → used
db.leave_balances.updateOne(
  { user_id: <userID>, leave_type: "annual_leave", year: 2026 },
  {
    $inc: { used_days: 3, pending_days: -3 },
    $set: { updated_at: new Date() }
  }
)

// ReleasePending — ปฏิเสธ: คืน pending กลับ
db.leave_balances.updateOne(
  { user_id: <userID>, leave_type: "annual_leave", year: 2026 },
  {
    $inc: { pending_days: -3 },
    $set: { updated_at: new Date() }
  }
)
```

### Collection: `leave_requests`

| Index | Fields | Type | วัตถุประสงค์ |
|---|---|---|---|
| `user_id_1` | `{ user_id: 1 }` | Normal | ค้นหาใบลาของพนักงานแต่ละคน |
| `status_1` | `{ status: 1 }` | Normal | ค้นหาใบลาตามสถานะ (pending queue สำหรับ Manager) |
| `user_id_1_start_date_1_end_date_1` | `{ user_id: 1, start_date: 1, end_date: 1 }` | **Compound** | ตรวจสอบวันลาซ้ำซ้อน (overlap check) |

```javascript
// ดูใบลาทั้งหมดของ user (เรียงใหม่สุดก่อน + pagination)
db.leave_requests.find({ user_id: <userID> })
  .sort({ created_at: -1 })
  .skip(0).limit(10)

// ดูใบลาที่รออนุมัติ (สำหรับ Manager — FIFO เรียงเก่าสุดก่อน)
db.leave_requests.find({ status: "pending" })
  .sort({ created_at: 1 })
  .skip(0).limit(10)

// Overlap check (ดู section Overlap Rule)
db.leave_requests.countDocuments({
  user_id: <userID>,
  status: { $in: ["pending", "approved"] },
  start_date: { $lte: <new_end> },
  end_date:   { $gte: <new_start> }
})

// Atomic approve — อัปเดตเฉพาะเมื่อสถานะยังเป็น pending (CAS)
db.leave_requests.replaceOne(
  { _id: <requestID>, status: "pending" },
  <updatedDocument>
)
```

---

## 💡 เหตุผลในการออกแบบ

### ทำไมใช้ Hexagonal Architecture?

แยก business logic ออกจาก framework และ database อย่างชัดเจน ทำให้:
- **ทดสอบง่าย** — ทดสอบ service ได้โดยไม่ต้องเชื่อมต่อ database จริง (ใช้ mock)
- **เปลี่ยน framework ได้** — ถ้าจะเปลี่ยนจาก Fiber เป็น Gin แก้แค่ adapters
- **เปลี่ยน database ได้** — ถ้าจะเปลี่ยนจาก MongoDB เป็น PostgreSQL แก้แค่ repositories
- **บังคับด้วย lint** — ใช้ `depguard` ใน `.golangci.yml` ตรวจสอบว่าไม่มี layer ไหนพึ่งพาผิดทิศทาง

### ทำไมใช้ UUID แทน MongoDB ObjectID?

- UUID เป็นมาตรฐานสากล ไม่ผูกกับฐานข้อมูลใดฐานข้อมูลหนึ่ง
- สร้างฝั่ง application ได้เลย ไม่ต้องรอ database
- ใช้เป็น domain type ได้โดยไม่ต้อง import database driver ใน domain layer

### ทำไมใช้ระบบ 3 ขั้นตอน (Reserve → Confirm/Release)?

การจัดการยอดวันลาใช้ระบบ `pending_days` แยกออกจาก `used_days`:

```
ยื่นใบลา  →  ReservePending  (เพิ่ม pending_days แบบ atomic)
อนุมัติ   →  ConfirmPending  (ย้าย pending_days → used_days)
ปฏิเสธ   →  ReleasePending  (คืน pending_days กลับ)
```

**ข้อดี:** ป้องกันพนักงานยื่นลาเกินโควตาขณะรออนุมัติ เช่น มีสิทธิ์ลา 15 วัน ยื่นไป 10 วัน ยื่นอีก 10 วันจะไม่ได้เพราะ pending_days ถูกนับรวมแล้ว

### ป้องกัน Race Condition

ระบบใช้เทคนิค Atomic CAS (Compare-And-Swap) ป้องกันกรณี Manager หลายคน approve/reject ใบลาเดียวกันพร้อมกัน โดยอัปเดตสถานะใบลาผ่าน MongoDB single-document atomicity — ไม่ต้องใช้ database transaction หรือ Replica Set

### ทำไมไม่มี Register Endpoint?

ระบบนี้ออกแบบให้ผู้ใช้ถูกสร้างผ่าน seed script หรือ admin tool — ไม่เปิดให้ลงทะเบียนเองผ่าน API เพราะ:
- ระบบจัดการลาเป็นระบบภายในองค์กร ไม่ใช่ระบบ public
- การสร้างผู้ใช้ควรผ่านกระบวนการ HR ไม่ใช่ให้ลงทะเบียนเอง

---

## 🔒 ความปลอดภัย

| มาตรการ | รายละเอียด |
|---------|-----------|
| **bcrypt** (cost 12) | เข้ารหัสรหัสผ่าน — ใช้เวลา ~250ms ต่อครั้ง ป้องกัน brute force |
| **JWT HS256** | token มี expiration กำหนดได้ผ่าน environment variable |
| **Rate Limiting** | จำกัด 10 requests/นาที ต่อ IP สำหรับ endpoint ยืนยันตัวตน |
| **Security Headers** | ป้องกัน XSS, Clickjacking, MIME sniffing (HSTS, CSP, X-Frame-Options ฯลฯ) |
| **Input Validation** | ตรวจสอบข้อมูลขาเข้าทุก endpoint ด้วย validator v10 |
| **Body Size Limit** | จำกัดขนาด request body ที่ 1MB |
| **CORS** | ตั้งค่า Cross-Origin Resource Sharing |
| **Non-root Docker** | Container รันด้วย user ที่ไม่ใช่ root |

---

## 🧪 การทดสอบและคุณภาพโค้ด

### รันทดสอบ

```bash
# รัน unit tests ทั้งหมด
go test ./... -v

# รัน tests พร้อม coverage
go test ./... -cover
```

**ครอบคลุม logic สำคัญ:**
- ✅ คำนวณยอดวันลา (หัก, คืน, ตรวจสอบยอดเพียงพอ, pending days)
- ✅ สร้างใบลา, อนุมัติ, ปฏิเสธ (สถานะเปลี่ยนถูกต้อง)
- ✅ คำนวณจำนวนวันลา
- ✅ ตรวจสอบ role, ประเภทการลา, สถานะใบลา
- ✅ Login สำเร็จ, อีเมลไม่ถูกต้อง, รหัสผ่านผิด
- ✅ ยื่นใบลา — ประเภทไม่ถูกต้อง, วันที่ไม่ถูกต้อง, วันลาซ้ำซ้อน, ยอดไม่พอ
- ✅ อนุมัติ/ปฏิเสธตัวเองไม่ได้, ห้ามอนุมัติใบลาที่ไม่ใช่สถานะ pending
- ✅ Rollback เมื่อสร้างใบลาในฐานข้อมูลล้มเหลว

### ตรวจสอบคุณภาพโค้ด

```bash
# ตรวจสอบ lint (รวม architecture dependency rules)
golangci-lint run

# Build ทั้งโปรเจค
go build ./...
```

---

## ⚠️ ข้อจำกัดที่ทราบ

| ข้อจำกัด | รายละเอียด | แนวทางปรับปรุง |
|---|---|---|
| ไม่ตัดวันหยุด | นับ calendar days รวมเสาร์-อาทิตย์และวันหยุดนักขัตฤกษ์ | เพิ่ม holiday list + business day calculation |
| ไม่รองรับ half-day | ลาได้เฉพาะเต็มวัน | เพิ่ม field `period` (morning/afternoon) ใน request |
| Timezone เดียว (UTC) | ไม่รองรับ timezone ของผู้ใช้แต่ละคน | เพิ่ม timezone setting ต่อ user |
| ไม่มี Register API | สร้างผู้ใช้ผ่าน seed script เท่านั้น | เพิ่ม admin endpoint สำหรับจัดการผู้ใช้ |
| ไม่มี Cancel ใบลา | พนักงานยกเลิกใบลาที่ยื่นไปแล้วไม่ได้ | เพิ่ม cancel endpoint + คืนยอด pending |
| ไม่มี Notification | ไม่แจ้งเตือนเมื่อมีใบลาใหม่หรือถูก approve/reject | เพิ่ม email/webhook notification |
