package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the API Gateway
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	RateLimit    RateLimitConfig
	Services     ServicesConfig
	Log          LogConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Environment     string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int
	MinConns int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	RefreshHours    int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled          bool
	RequestsPerMinute int
}

// ServicesConfig holds backend service addresses
type ServicesConfig struct {
	OrchestratorGRPC string
	StorageGRPC      string
	WebSocketHTTP    string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			Environment:     getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "cloudscan"),
			Password: getEnv("DB_PASSWORD", "changeme"),
			Database: getEnv("DB_NAME", "cloudscan_gateway"),
			SSLMode:  getEnv("DB_SSLMODE", "prefer"),
			MaxConns: getIntEnv("DB_MAX_CONNS", 25),
			MinConns: getIntEnv("DB_MIN_CONNS", 5),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "changeme-secret-key"),
			ExpirationHours: getIntEnv("JWT_EXPIRATION_HOURS", 24),
			RefreshHours:    getIntEnv("JWT_REFRESH_HOURS", 168), // 7 days
		},
		RateLimit: RateLimitConfig{
			Enabled:          getBoolEnv("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getIntEnv("RATE_LIMIT_RPM", 60),
		},
		Services: ServicesConfig{
			OrchestratorGRPC: getEnv("ORCHESTRATOR_GRPC_URL", "orchestrator:9999"),
			StorageGRPC:      getEnv("STORAGE_GRPC_URL", "storage:9998"),
			WebSocketHTTP:    getEnv("WEBSOCKET_HTTP_URL", "http://websocket:9090"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return "host=" + c.Host + " port=" + c.Port + " user=" + c.User + " password=" + c.Password + " dbname=" + c.Database + " sslmode=" + c.SSLMode
}

// GetRedisAddr returns Redis connection address
func (c *RedisConfig) GetAddr() string {
	return c.Host + ":" + c.Port
}