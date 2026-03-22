package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"golang-app/internal/config"
)

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	// Khởi tạo redis client với thông tin cấu hình
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Kiểm tra kết nối
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis at %s:%d - error: %v", cfg.Redis.Host, cfg.Redis.Port, err)
		return nil, err
	}

	log.Printf("Connected to Redis successfully at %s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return rdb, nil
}
