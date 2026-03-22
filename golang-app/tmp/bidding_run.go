package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const baseURL = "http://localhost:8080/api/v1"
const wsURL = "ws://localhost:8080/api/v1/ws/auctions/1"

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type BidReq struct {
	Amount float64 `json:"amount"`
}

func main() {
	fmt.Println("🚀 BẮT ĐẦU CHẠY KỊCH BẢN ĐẤU GIÁ TỰ ĐỘNG...")
	fmt.Println("-------------------------------------------------")

	// 1. Đăng nhập lấy Token của Đại gia A (Bidder)
	tokenA := login("bidderA@example.com", "password123") // Đổi email/pass theo đúng DB của bạn nếu cần
	if tokenA == "" {
		// Dùng tạm token cứng nếu login lỗi do rớt DB
		fmt.Println("❌ Không lấy được token. Vui lòng đảm bảo Server đang bật và có user bidderA@example.com")
		// Trả về mock error để skip script nếu lỗi
		// return 
	}
	fmt.Println("✅ [Bước 1] Đăng nhập thành công, đã lấy được Token.")

	// 2. Kết nối WebSocket giả lập Tab 1 trong Postman
	fmt.Println("⏳ [Bước 2] Đang kết nối WebSocket tới kênh Auction 1...")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("❌ Lỗi kết nối WebSocket: %v", err)
	}
	defer conn.Close()
	fmt.Println("✅ [Bước 2] WebSocket đã kết nối THÀNH CÔNG!")

	// 3. Khởi tạo Goroutine (Luồng phụ) để liên tục LẮNG NGHE tin nhắn gửi từ Server về
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket đóng kết nối:", err)
				return
			}
			fmt.Printf("\n🔔 [WEBSOCKET LIVE BÁO CÁO]: Server vừa đẩy Event mới!\n")
			fmt.Printf("📦 Dữ liệu JSON mới nhất: %s\n\n", string(message))
		}
	}()

	// 4. Bắn API Đặt giá (Rest API HTTP) giống hệt lúc ấn Send trong Postman
	time.Sleep(1 * time.Second) // Chờ 1s cho WS ổn định
	fmt.Println("💸 [Bước 3] Tiến hành gửi Request Đặt giá: 16,000,000 đ...")
	placeBid(tokenA, 16000000)

	// Chờ 2 giây để nhận dữ liệu Broadcast bay ngược về, sau đó tắt tool.
	time.Sleep(2 * time.Second)
	fmt.Println("🏁 CHƯƠNG TRÌNH TEST ĐÃ HOÀN TẤT!")
}

// Hàm phụ trợ Login lấy JWT Token
func login(email, password string) string {
	reqBody, _ := json.Marshal(LoginReq{Email: email, Password: password})
	resp, err := http.Post(baseURL+"/users/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()
	
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if data, ok := res["data"].(map[string]interface{}); ok {
		return data["token"].(string)
	}
	return ""
}

// Hàm phụ trợ Bắn API Bid
func placeBid(token string, amount float64) {
	reqBody, _ := json.Marshal(BidReq{Amount: amount})
	req, _ := http.NewRequest("POST", baseURL+"/auctions/1/bids", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Lỗi gửi Bid: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   -> Phản hồi từ Server HTTP (%d): %s\n", resp.StatusCode, string(body))
}
