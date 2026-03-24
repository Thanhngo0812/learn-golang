package repository

import (
	"database/sql"
	"golang-app/internal/model"
)

// UserRepository định nghĩa các hành vi tương tác với bảng User
type UserRepository interface {
	CreateUser(user *model.User) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository là hàm khởi tạo (Constructor)
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Implement hàm CreateUser
func (r *userRepository) CreateUser(user *model.User) error {
	query := "INSERT INTO users (name, phone) VALUES ($1, $2) RETURNING id"
	// Thực thi query và lấy ID vừa tạo gán ngược lại vào user.ID
	err := r.db.QueryRow(query, user.Name, user.Phone).Scan(&user.ID)
	return err
}
