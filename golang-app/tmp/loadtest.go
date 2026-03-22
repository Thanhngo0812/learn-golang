package main

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := "http://localhost:8080/api/v1/auctions/1"
	fmt.Printf("Bắt đầu thực hiện bài test tải cho Rate Limiter.\n")
	fmt.Printf("Mục tiêu: Bắn 250 requests CÙNG 1 LÚC vào %s\n\n", url)

	// Các biến lưu trữ thống kê (dùng atomic để goroutine đếm an toàn)
	var successCount int32       // HTTP 200
	var tooManyRequestsCount int32 // HTTP 429
	var otherCount int32         // Lỗi khác

	totalRequests := 250
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	// Dùng channel để "cầm nhịp" cho tất cả 250 goroutine xuất phát ĐỒNG THỜI
	startSignal := make(chan struct{})

	// Cấu hình Transport để không bị thắt cổ chai ở phía Client khi gửi 250 requests đến localhost
	tr := &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 500,
		MaxConnsPerHost:     500,
	}
	client := http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	// Khởi tạo 250 Goroutines (luồng xử lý đồng thời)
	for i := 0; i < totalRequests; i++ {
		go func(reqID int) {
			defer wg.Done()
			
			// Goroutine sinh ra sẽ nằm đợi tín hiệu ở đây
			<-startSignal

			resp, err := client.Get(url)
			if err != nil {
				atomic.AddInt32(&otherCount, 1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				atomic.AddInt32(&successCount, 1)
			} else if resp.StatusCode == http.StatusTooManyRequests {
				atomic.AddInt32(&tooManyRequestsCount, 1)
			} else {
				atomic.AddInt32(&otherCount, 1)
			}
		}(i)
	}

	// Đợi 1 chút xíu để tất cả goroutines kịp sinh ra và đợi sẵn sàng.
	time.Sleep(500 * time.Millisecond)

	// Kéo cò! Đóng channel -> tất cả 250 goroutines bắt đầu gọi API cùng 1 tích tắc.
	fmt.Println("🚀 BẮN REQUESTS ĐỒNG THỜI...")
	start := time.Now()
	close(startSignal)

	// Đợi tất cả request chạy xong
	wg.Wait()
	duration := time.Since(start)

	// In kết quả
	fmt.Println("======================================")
	fmt.Printf("⏱️  Thời gian test: %v\n", duration)
	fmt.Printf("✅ Số requests THÀNH CÔNG (200 OK): %d\n", successCount)
	fmt.Printf("🛑 Số requests BỊ CHẶN (429 Too Many Requests): %d\n", tooManyRequestsCount)
	fmt.Printf("⚠️ Lỗi khác: %d\n", otherCount)
	fmt.Println("======================================")
	
	if successCount == 200 && tooManyRequestsCount == 50 {
		fmt.Println("🎉 KẾT LUẬN: RATE LIMITER HOẠT ĐỘNG CHÍNH XÁC!!!")
		fmt.Println("-> Chỉ có tối đa 200 kết nối được xử lý, 50 kết nối dư thừa bị từ chối ngay lập tức.")
	} else if successCount > 0 && tooManyRequestsCount > 0 {
		fmt.Println("👍 KẾT LUẬN: RATE LIMITER CÓ HOẠT ĐỘNG (Đã thấy HTTP 429).")
		fmt.Println("-> Chú ý: Vì máy tính xử lý API quá nhanh hoặc CPU rảnh rỗi chia luồng, số lượng có thể sai số so với mức 200/50 lý thuyết ở local.")
	} else {
		fmt.Println("❌ RATE LIMITER CHƯA HOẠT ĐỘNG NHƯ MONG ĐỢI.")
	}
}
