package validator

import (
	"errors"
	"github.com/go-playground/validator/v10"
)

// Struct định nghĩa format lỗi trả về cho từng field
type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Hàm này biến đổi error gốc thành list lỗi đẹp (nếu đúng là lỗi validation)
func FormatError(err error) []ValidationErrorResponse {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]ValidationErrorResponse, len(ve))
		for i, fe := range ve {
			out[i] = ValidationErrorResponse{
				Field:   fe.Field(), // Tên field (VD: Email)
				Message: msgForTag(fe), // Hàm dịch lỗi
			}
		}
		return out
	}
	return nil // Không phải lỗi validation
}

// Hàm phụ trợ để dịch message (Custom message ở đây)
func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Thông tin này là bắt buộc"
	case "email":
		return "Email không đúng định dạng"
	case "len":
		return "Độ dài phải đúng " + fe.Param() + " ký tự"
	case "min":
		return "Độ dài tối thiểu là " + fe.Param() + " ký tự"
	case "max":
		return "Độ dài tối đa là " + fe.Param() + " ký tự"
	case "oneof":
        return "Giá trị phải là một trong các loại: " + fe.Param()
    
	}
	
	return fe.Error() // Mặc định trả về lỗi gốc nếu chưa dịch
}