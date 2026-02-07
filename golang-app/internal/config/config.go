package config

import (
	"log"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	App    AppConfig    `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"` // Thêm dòng này

}

type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	Timeout int    `mapstructure:"timeout"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}
func LoadConfig() (*Config, error) {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 1. Thay thế dấu chấm (.) bằng dấu gạch dưới (_)
	// Để viper hiểu rằng: database.host (trong yaml) <==> DATABASE_HOST (biến môi trường)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 2. Đọc biến môi trường
	viper.AutomaticEnv()

	// 3. Đọc file config
	if err := viper.ReadInConfig(); err != nil {
		
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	log.Printf("Config loaded successfully: %+v", config)
	return &config, nil
}
