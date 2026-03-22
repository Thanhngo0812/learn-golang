package response

import (
	"github.com/gin-gonic/gin"
	"golang-app/pkg/apperror"
	"golang-app/pkg/validator" // Import package validator vừa tạo
	"net/http"
)

// SuccessResponse đại diện cho phản hồi thành công
type SuccessResponse struct {
	Status string      `json:"status" example:"success"`
	Data   interface{} `json:"data"`
}

// ErrorResponse đại diện cho phản hồi lỗi
type ErrorResponse struct {
	Status  string      `json:"status" example:"error"`
	Message string      `json:"message" example:"Đã có lỗi xảy ra"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Trả về thành công
func Success(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{
		"status": "success",
		"data":   data,
	})
}

// Trả về lỗi
// pkg/response/response.go

func Error(c *gin.Context, err error) {
    // 1. Ưu tiên kiểm tra lỗi Validation trước
    // Hàm validator.FormatError này chúng ta đã viết ở câu trả lời trước
    if validationErrors := validator.FormatError(err); validationErrors != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status":  "error",
            "message": "Dữ liệu đầu vào không hợp lệ",
            "errors":  validationErrors, // Trả về mảng lỗi đẹp (Field, Message)
        })
        return
    }

    // 2. Kiểm tra lỗi AppError (Lỗi logic do bạn tự định nghĩa, ví dụ: Sai pass, Email trùng...)
    if appErr, ok := err.(*apperror.AppError); ok {
        c.JSON(appErr.Code, gin.H{
            "status":  "error",
            "message": appErr.Message,
        })
        return
    }

    // 3. Các lỗi còn lại (Lỗi cú pháp JSON, Lỗi DB connection...)
    c.JSON(http.StatusInternalServerError, gin.H{
        "status":  "error",
        "message": "Đã có lỗi không mong muốn xảy ra",
    })
}