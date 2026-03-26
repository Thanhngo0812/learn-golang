package middleware

import (
	"golang-app/pkg/apperror"
	"golang-app/pkg/response"
	"github.com/gin-gonic/gin"
)

// ConcurrencyLimiter sử dụng một buffered channel (áp dụng nguyên lý Semaphore)
// để giới hạn số lượng request được xử lý đồng thời tới một API tải nặng.
// Nếu giới hạn đã đạt mức tối đa, hệ thống sẽ thực hiện Fail-fast, 
// ngay lập tức từ chối request mới và trả về lỗi 429 Too Many Requests.
func ConcurrencyLimiter(maxConcurrent int) gin.HandlerFunc {
	// Khởi tạo một buffered channel (kênh đệm) với kích thước bằng maxConcurrent
	sema := make(chan struct{}, maxConcurrent)

	return func(c *gin.Context) {
		select {
		case sema <- struct{}{}: // Cố gắng đẩy một tín hiệu rỗng (token) vào channel
			// Thành công lấy được semaphore (khe xử lý).
			// Đảm bảo token sẽ được giải phóng ra khỏi channel khi request hoàn thành.
			defer func() { <-sema }()
			
			// Request được phép đi tiếp tới hàm thực thi (Tầng Controller)
			c.Next()
		default:
			// Channel đã đầy chữ (đạt mức maxConcurrent). Chặn đứng và Fail-fast.
			response.Error(c, apperror.NewTooManyRequests(nil, "Server đang quá tải, vui lòng thử lại"))
			c.Abort()
			return
		}
	}
}
