package apperror

import "net/http"

type AppError struct {
	Code    int    `json:"code"`    // HTTP Status Code
	Message string `json:"message"` // Thông báo lỗi
	Raw     error  `json:"-"`       // Lỗi gốc (để log, không trả về client)
}

func (e *AppError) Error() string {
	return e.Message
}

// 400: Lỗi dữ liệu đầu vào
func NewBadRequest(err error, msg string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: msg,
		Raw:     err,
	}
}

// 409: Lỗi trùng lặp (Conflict)
func NewConflict(err error, msg string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: msg,
		Raw:     err,
	}
}

func NewNotFound(err error, msg string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound, // 404
		Message: msg,
		Raw:     err,
	}
}
// 401: Lỗi xác thực
func NewUnauthorized(err error, msg string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: msg,
		Raw:     err,
	}
}

// 403: Lỗi cấm truy cập (Forbidden)
func NewForbidden(err error, msg string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: msg,
		Raw:     err,
	}
}

// 429: Lỗi quá trình request (Too Many Requests)
func NewTooManyRequests(err error, msg string) *AppError {
	return &AppError{
		Code:    http.StatusTooManyRequests,
		Message: msg,
		Raw:     err,
	}
}

// 500: Lỗi hệ thống
func NewInternal(err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Lỗi hệ thống nội bộ",
		Raw:     err,
	}
}