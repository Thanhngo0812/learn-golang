package utils

import (
	"strconv"
	"strings"
	"golang-app/pkg/apperror" 
	"github.com/gin-gonic/gin"
)

// GetIDFromParam: Lấy ID từ URL và chuyển sang int, tự động wrap lỗi vào apperror
func GetIDFromParam(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	
	if err != nil {
		// Dùng chính cái apperror đã định nghĩa trong pkg
		return 0, apperror.NewBadRequest(err, "ID phải là định dạng số nguyên")
	}
	
	if id <= 0 {
		return 0, apperror.NewBadRequest(nil, "ID không hợp lệ (phải > 0)")
	}

	return id, nil
}

// GetIntQuery: Lấy một giá trị integer từ query parameter, hỗ trợ giá trị mặc định
func GetIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue, err
	}
	return val, nil
}

// GetUintQuery: Lấy một giá trị uint từ query parameter, hỗ trợ giá trị mặc định
func GetUintQuery(c *gin.Context, key string, defaultValue uint) (uint, error) {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return defaultValue, err
	}
	return uint(val), nil
}

// ParseIntSlice: Phân tách chuỗi comma-separated (1,2,3) thành slice []int
func ParseIntSlice(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			res = append(res, id)
		}
	}
	return res
}