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
	Redis  RedisConfig  `mapstructure:"redis"`
	Cloudinary CloudinaryConfig `mapstructure:"cloudinary"`
}

type CloudinaryConfig struct {
	CloudName    string `mapstructure:"cloud_name"`
	APIKey       string `mapstructure:"api_key"`
	APISecret    string `mapstructure:"api_secret"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ServerConfig struct {
	Port    int    `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	Timeout int    `mapstructure:"timeout"`
}

type AppConfig struct {
	Name          string `mapstructure:"name"`
	Version       string `mapstructure:"version"`
	JWTSecret     string `mapstructure:"jwt_secret"`
	JWTExpiration int    `mapstructure:"jwt_expiration"`
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
	// 1. Load .env file using a separate viper instance to avoid type conflicts
	envViper := viper.New()
	envViper.SetConfigFile(".env")
	envViper.SetConfigType("env")
	if err := envViper.ReadInConfig(); err == nil {
		for _, key := range envViper.AllKeys() {
			val := envViper.GetString(key)
			// Map specific .env keys to viper expected nested keys
			switch key {
			case "cloudinary_cloud_name": viper.Set("cloudinary.cloud_name", val)
			case "cloudinary_api_key": viper.Set("cloudinary.api_key", val)
			case "cloudinary_api_secret": viper.Set("cloudinary.api_secret", val)
			default: viper.Set(key, val)
			}
		}
		log.Println(".env file loaded successfully")
	}

	// 2. Setup main viper for config.yaml
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 4. Bind Nested Environment Variables
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "DATABASE_DBNAME")
	viper.BindEnv("cloudinary.cloud_name", "CLOUDINARY_CLOUD_NAME")
	viper.BindEnv("cloudinary.api_key", "CLOUDINARY_API_KEY")
	viper.BindEnv("cloudinary.api_secret", "CLOUDINARY_API_SECRET")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	log.Printf("Config loaded successfully: %+v", config)
	return &config, nil
}
