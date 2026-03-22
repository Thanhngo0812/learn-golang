package db

import (
    "fmt"
    "golang-app/internal/config"
    "log"

    "gorm.io/driver/postgres" // Sử dụng driver của gorm
    "gorm.io/gorm"           // Import GORM
)

// Đổi kiểu trả về từ *sql.DB thành *gorm.DB
func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.DBName,
        cfg.Database.SSLMode,
    )

    // Sử dụng gorm.Open thay vì sql.Open
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // GORM đã tự động Ping khi kết nối, nhưng nếu bạn muốn chắc chắn:
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
    }
    
    if err = sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Println("✅ Connected to PostgreSQL via GORM successfully!")
    return db, nil
}