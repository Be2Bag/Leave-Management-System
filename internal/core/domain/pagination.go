package domain

const (
	DefaultPage     = 1   // หน้าเริ่มต้น
	DefaultPageSize = 10  // จำนวนรายการต่อหน้าเริ่มต้น
	MaxPageSize     = 100 // จำนวนรายการต่อหน้าสูงสุด
)

// PaginationParams พารามิเตอร์สำหรับแบ่งหน้า
type PaginationParams struct {
	Page     int // หน้าที่ต้องการ (เริ่มจาก 1)
	PageSize int // จำนวนรายการต่อหน้า
}

// NewPaginationParams สร้าง PaginationParams พร้อม normalize ค่า
func NewPaginationParams(page, pageSize int) PaginationParams {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return PaginationParams{Page: page, PageSize: pageSize}
}

func (p PaginationParams) Offset() int64 {
	return int64((p.Page - 1) * p.PageSize)
}
func (p PaginationParams) Limit() int64 {
	return int64(p.PageSize)
}

type PaginatedResult[T any] struct {
	Items      []T   // รายการข้อมูล
	Total      int64 // จำนวนรายการทั้งหมด
	Page       int   // หน้าปัจจุบัน
	PageSize   int   // จำนวนรายการต่อหน้า
	TotalPages int   // จำนวนหน้าทั้งหมด
}

// NewPaginatedResult สร้าง PaginatedResult พร้อมคำนวณจำนวนหน้าทั้งหมด
func NewPaginatedResult[T any](items []T, total int64, params PaginationParams) *PaginatedResult[T] {
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}
	if items == nil {
		items = []T{}
	}
	return &PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}
}
