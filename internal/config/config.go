package config

import (
	"time"
)

// Config 主配置结构
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	MinIO    MinIOConfig    `yaml:"minio"`
	JWT      JWTConfig      `yaml:"jwt"`
	Redis    RedisConfig    `yaml:"redis"`
	SMS      SMSConfig      `yaml:"sms"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name           string   `yaml:"name"`
	Env            string   `yaml:"env"`
	Port           int      `yaml:"port"`
	Debug          bool     `yaml:"debug"`
	AllowedOrigins []string `yaml:"cors.allowed_origins"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver                string        `yaml:"driver"`
	Host                  string        `yaml:"host"`
	Port                  int           `yaml:"port"`
	Username              string        `yaml:"username"`
	Password              string        `yaml:"password"`
	DBName                string        `yaml:"dbname"`
	MaxOpenConnections    int           `yaml:"max_open_connections"`
	MaxIdleConnections    int           `yaml:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `yaml:"connection_max_lifetime"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	JwtSecret        string `yaml:"jwt_secret"`
	ExpirationHours  int    `yaml:"expiration_hours"`
	RefreshSecret    string `yaml:"refresh_secret"`
	SessionKeySecret string `yaml:"session_key_secret"`
}

// MilvusConfig Milvus连接配置
type MilvusConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	DBName string `yaml:"db_name"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// SMSConfig 阿里云短信配置
type SMSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	SignName        string `yaml:"sign_name"`
	TemplateCode    string `yaml:"template_code"`
}
