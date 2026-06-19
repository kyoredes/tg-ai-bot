package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host           string
	Port           string
	Debug          bool
	LoggingMode    string
	Timeout        time.Duration
	ServerTokenTTL time.Duration
}

func NewConfig() *Config {
	host := viper.GetString("HOST")
	port := viper.GetString("PORT")
	debug := viper.GetBool("DEBUG")
	loggingMode := viper.GetString("LOGGING_MODE")
	timeout := viper.GetDuration("TIMEOUT")
	serverTokenTTL := viper.GetDuration("SERVER_TOKEN_TTL")

	return &Config{
		Host:           host,
		Port:           port,
		Debug:          debug,
		LoggingMode:    loggingMode,
		Timeout:        timeout,
		ServerTokenTTL: serverTokenTTL,
	}
}

type AuthConfig struct {
	AuthHost       string
	AuthPort       string
	AuthAPIKey     string
	AuthTimeout    int
	AuthRetryCount int
}

func NewAuthConfig() *AuthConfig {
	authHost := viper.GetString("AUTH_HOST")
	authPort := viper.GetString("AUTH_PORT")
	authAPIKey := viper.GetString("AUTH_API_KEY")
	authTimeout := viper.GetInt("AUTH_TIMEOUT")
	authRetryCount := viper.GetInt("AUTH_RETRY_COUNT")

	return &AuthConfig{
		AuthHost:       authHost,
		AuthPort:       authPort,
		AuthAPIKey:     authAPIKey,
		AuthTimeout:    authTimeout,
		AuthRetryCount: authRetryCount,
	}
}

type RedisConfig struct {
	Host     string
	Port     string
	DB       int
	Password string
}

func NewRedisConfig() *RedisConfig {
	host := viper.GetString("REDIS_HOST")
	port := viper.GetString("REDIS_PORT")
	db := viper.GetInt("REDIS_DB")
	password := viper.GetString("REDIS_PASSWORD")

	return &RedisConfig{
		Host:     host,
		Port:     port,
		DB:       db,
		Password: password,
	}
}

type DevConfig struct {
	CommonPubKey string
}

func NewDevConfig() *DevConfig {
	commonPubKey := viper.GetString("COMMON_PUB_KEY")

	return &DevConfig{
		CommonPubKey: commonPubKey,
	}
}

type SubConfig struct {
	SubHost       string
	SubPort       string
	SubAPIKey     string
	SubTimeout    int
	SubRetryCount int
}

func NewSubConfig() *SubConfig {
	subHost := viper.GetString("SUB_HOST")
	subPort := viper.GetString("SUB_PORT")
	subAPIKey := viper.GetString("SUB_API_KEY")
	subTimeout := viper.GetInt("SUB_TIMEOUT")
	subRetryCount := viper.GetInt("SUB_RETRY_COUNT")

	return &SubConfig{
		SubHost:       subHost,
		SubPort:       subPort,
		SubAPIKey:     subAPIKey,
		SubTimeout:    subTimeout,
		SubRetryCount: subRetryCount,
	}
}

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8000")

	viper.SetDefault("DEBUG", false)
	viper.SetDefault("LOGGING_MODE", "text")
	viper.SetDefault("TIMEOUT", "10s")

	viper.SetDefault("AUTH_HOST", "localhost")
	viper.SetDefault("AUTH_PORT", "8099")
	viper.SetDefault("AUTH_API_KEY", "api_key")
	viper.SetDefault("AUTH_TIMEOUT", 10)
	viper.SetDefault("AUTH_RETRY_COUNT", 3)

	viper.SetDefault("SUB_HOST", "localhost")
	viper.SetDefault("SUB_PORT", "8099")
	viper.SetDefault("SUB_API_KEY", "api_key")
	viper.SetDefault("SUB_TIMEOUT", 10)
	viper.SetDefault("SUB_RETRY_COUNT", 3)

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_PASSWORD", "123")

	viper.SetDefault("COMMON_PUB_KEY", "secret")
}
