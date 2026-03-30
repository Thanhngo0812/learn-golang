package controller

import (
	customWebSocket "golang-app/pkg/websocket"
	"golang-app/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Cho phép tất cả origins (CORS) cho việc test local, trên production cần fix lại
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSController struct {
	hub *customWebSocket.Hub
}

func NewWSController(hub *customWebSocket.Hub) *WSController {
	return &WSController{hub: hub}
}

// ServeWS nâng cấp (upgrade) HTTP req thành Websocket Connection cho một phòng đấu giá cụ thể
func (c *WSController) ServeWS(ctx *gin.Context) {
	auctionID, _ := utils.GetIDFromParam(ctx)
	// UserID có thể có hoặc không (người xem vãng lai)
	var userID uint
	if u, exists := ctx.Get("userID"); exists {
		userID = u.(uint)
	}

	conn, wsErr := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if wsErr != nil {
		return
	}

	client := customWebSocket.NewClient(c.hub, conn, uint(auctionID), userID)
	c.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

// ServeNotifications dành cho kết nối toàn cục để nhận thông báo real-time
func (c *WSController) ServeNotifications(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(uint)

	conn, wsErr := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if wsErr != nil {
		return
	}

	// AuctionID = 0 đại diện cho kết nối toàn cục (chỉ nhận thông báo)
	client := customWebSocket.NewClient(c.hub, conn, 0, userID)
	c.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
