package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host        string
	Port        string
	LoggingMode string
	Timeout     time.Duration
	KafkaBrokers []string
}

func NewConfig() *Config {
	return &Config{
		Host:         viper.GetString("HOST"),
		Port:         viper.GetString("PORT"),
		LoggingMode:  viper.GetString("LOGGING_MODE"),
		Timeout:      viper.GetDuration("TIMEOUT"),
		KafkaBrokers: parseKafkaBrokers(),
	}
}

func parseKafkaBrokers() []string {
	raw := viper.GetString("KAFKA_BROKERS")
	if raw == "" {
		return []string{"localhost:9092"}
	}
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		if broker := strings.TrimSpace(part); broker != "" {
			brokers = append(brokers, broker)
		}
	}
	if len(brokers) == 0 {
		return []string{"localhost:9092"}
	}
	return brokers
}

type AuthConfig struct {
	AuthHost     string
	AuthGRPCPort string
	AuthTimeout  int
}

func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		AuthHost:     viper.GetString("AUTH_HOST"),
		AuthGRPCPort: viper.GetString("AUTH_GRPC_PORT"),
		AuthTimeout:  viper.GetInt("AUTH_TIMEOUT"),
	}
}

type DevConfig struct {
	CommonPubKey string
}

func NewDevConfig() *DevConfig {
	return &DevConfig{
		CommonPubKey: viper.GetString("COMMON_PUB_KEY"),
	}
}

type SubConfig struct {
	SubHost     string
	SubGRPCPort string
}

func NewSubConfig() *SubConfig {
	return &SubConfig{
		SubHost:     viper.GetString("SUB_HOST"),
		SubGRPCPort: viper.GetString("SUB_GRPC_PORT"),
	}
}

type AIConfig struct {
	AIHost     string
	AIGRPCPort string
}

func NewAIConfig() *AIConfig {
	return &AIConfig{
		AIHost:     viper.GetString("AI_HOST"),
		AIGRPCPort: viper.GetString("AI_GRPC_PORT"),
	}
}

type AdminConfig struct {
	Username   string
	Password   string
	JWTSecret  string
	CORSOrigin string
}

type ThrottleConfig struct {
	Enabled     bool
	Limit       int
	Window      time.Duration
	ChatLimit   int
	ChatWindow  time.Duration
	LoginLimit  int
	LoginWindow time.Duration
}

func NewThrottleConfig() *ThrottleConfig {
	return &ThrottleConfig{
		Enabled:     viper.GetBool("THROTTLE_ENABLED"),
		Limit:       viper.GetInt("THROTTLE_LIMIT"),
		Window:      viper.GetDuration("THROTTLE_WINDOW"),
		ChatLimit:   viper.GetInt("THROTTLE_CHAT_LIMIT"),
		ChatWindow:  viper.GetDuration("THROTTLE_CHAT_WINDOW"),
		LoginLimit:  viper.GetInt("THROTTLE_LOGIN_LIMIT"),
		LoginWindow: viper.GetDuration("THROTTLE_LOGIN_WINDOW"),
	}
}

func NewAdminConfig() *AdminConfig {
	return &AdminConfig{
		Username:   viper.GetString("ADMIN_USERNAME"),
		Password:   viper.GetString("ADMIN_PASSWORD"),
		JWTSecret:  viper.GetString("ADMIN_JWT_SECRET"),
		CORSOrigin: viper.GetString("ADMIN_CORS_ORIGIN"),
	}
}

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8000")
	viper.SetDefault("LOGGING_MODE", "text")
	viper.SetDefault("TIMEOUT", "120s")

	viper.SetDefault("THROTTLE_ENABLED", true)
	viper.SetDefault("THROTTLE_LIMIT", 120)
	viper.SetDefault("THROTTLE_WINDOW", "60s")
	viper.SetDefault("THROTTLE_CHAT_LIMIT", 10)
	viper.SetDefault("THROTTLE_CHAT_WINDOW", "60s")
	viper.SetDefault("THROTTLE_LOGIN_LIMIT", 10)
	viper.SetDefault("THROTTLE_LOGIN_WINDOW", "60s")

	viper.SetDefault("AUTH_HOST", "localhost")
	viper.SetDefault("AUTH_GRPC_PORT", "50051")
	viper.SetDefault("AUTH_TIMEOUT", 10)

	viper.SetDefault("SUB_HOST", "localhost")
	viper.SetDefault("SUB_GRPC_PORT", "50052")

	viper.SetDefault("AI_HOST", "localhost")
	viper.SetDefault("AI_GRPC_PORT", "50053")

	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")

	viper.SetDefault("COMMON_PUB_KEY", "secret")

	viper.SetDefault("ADMIN_USERNAME", "admin")
	viper.SetDefault("ADMIN_PASSWORD", "changeme")
	viper.SetDefault("ADMIN_JWT_SECRET", "admin-jwt-secret-change-in-prod")
	viper.SetDefault("ADMIN_CORS_ORIGIN", "http://localhost:5173")
}
