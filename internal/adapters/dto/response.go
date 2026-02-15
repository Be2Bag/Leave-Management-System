package dto

type APIResponse struct {
	Data    any    `json:"data,omitempty"` // ข้อมูลที่ส่งกลับ (ถ้ามี)
	Message string `json:"message"`        // ข้อความอธิบาย
	Success bool   `json:"success"`        // สถานะความสำเร็จ
}

type ErrorResponse struct {
	Message string   `json:"message"`          // ข้อความอธิบายข้อผิดพลาด
	Errors  []string `json:"errors,omitempty"` // รายละเอียดข้อผิดพลาด (ถ้ามี)
	Success bool     `json:"success"`          // สถานะ (false เสมอ)
}

type PaginationMeta struct {
	Page       int   `json:"page"`        // หน้าปัจจุบัน
	PageSize   int   `json:"page_size"`   // จำนวนรายการต่อหน้า
	Total      int64 `json:"total"`       // จำนวนรายการทั้งหมด
	TotalPages int   `json:"total_pages"` // จำนวนหน้าทั้งหมด
}

type PaginatedAPIResponse struct {
	Data       any            `json:"data"`       // รายการข้อมูล
	Message    string         `json:"message"`    // ข้อความอธิบาย
	Pagination PaginationMeta `json:"pagination"` // ข้อมูล pagination
	Success    bool           `json:"success"`    // สถานะความสำเร็จ
}

func NewSuccessResponse(message string, data any) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func NewPaginatedResponse(message string, data any, page, pageSize int, total int64, totalPages int) PaginatedAPIResponse {
	return PaginatedAPIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

func NewErrorResponse(message string, errors ...string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}

func NewValidationErrorResponse(errors []string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "ข้อมูลไม่ผ่านการตรวจสอบ",
		Errors:  errors,
	}
}
