package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host        string
	Port        string
	LoggingMode string
	Timeout     time.Duration
}

func NewConfig() *Config {
	return &Config{
		Host:        viper.GetString("HOST"),
		Port:        viper.GetString("PORT"),
		LoggingMode: viper.GetString("LOGGING_MODE"),
		Timeout:     viper.GetDuration("TIMEOUT"),
	}
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

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8000")
	viper.SetDefault("LOGGING_MODE", "text")
	viper.SetDefault("TIMEOUT", "10s")

	viper.SetDefault("AUTH_HOST", "localhost")
	viper.SetDefault("AUTH_GRPC_PORT", "50051")
	viper.SetDefault("AUTH_TIMEOUT", 10)

	viper.SetDefault("SUB_HOST", "localhost")
	viper.SetDefault("SUB_GRPC_PORT", "50052")

	viper.SetDefault("COMMON_PUB_KEY", "secret")
}
