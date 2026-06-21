package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host                   string
	Port                   string
	Debug                  bool
	LoggingMode            string
	JWTSecretKey           string
	AccessTokenExpiration  int
	RefreshTokenExpiration int
	Timeout                time.Duration
}

type GRPCConfig struct {
	Host string
	Port string
}

type ThrottleConfig struct {
	Enabled bool
	Limit   int
	Window  time.Duration
}

func NewThrottleConfig() *ThrottleConfig {
	return &ThrottleConfig{
		Enabled: viper.GetBool("THROTTLE_ENABLED"),
		Limit:   viper.GetInt("THROTTLE_LIMIT"),
		Window:  viper.GetDuration("THROTTLE_WINDOW"),
	}
}

func NewGRPCConfig() *GRPCConfig {
	host := viper.GetString("GRPC_HOST")
	port := viper.GetString("GRPC_PORT")

	return &GRPCConfig{
		Host: host,
		Port: port,
	}
}

type DBConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	DBTimeout time.Duration
	DBDSN     string
}

func NewConfig() *Config {
	host := viper.GetString("HOST")
	port := viper.GetString("PORT")
	debug := viper.GetBool("DEBUG")
	loggingMode := viper.GetString("LOGGING_MODE")
	jwtSecretKey := viper.GetString("JWT_SECRET_KEY")
	accessTokenExpiration := viper.GetInt("ACCESS_TOKEN_EXPIRATION")
	refreshTokenExpiration := viper.GetInt("REFRESH_TOKEN_EXPIRATION")
	timeout := viper.GetDuration("TIMEOUT")

	return &Config{
		Host:                   host,
		Port:                   port,
		Debug:                  debug,
		LoggingMode:            loggingMode,
		JWTSecretKey:           jwtSecretKey,
		AccessTokenExpiration:  accessTokenExpiration,
		RefreshTokenExpiration: refreshTokenExpiration,
		Timeout:                timeout,
	}
}

func NewDBConfig() *DBConfig {
	host := viper.GetString("DB_HOST")
	port := viper.GetInt("DB_PORT")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")
	dbName := viper.GetString("DB_NAME")
	dbTimeout := viper.GetDuration("DB_TIMEOUT")
	dbSSL := viper.GetString("DB_SSL")
	DBDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbName, dbSSL)

	return &DBConfig{
		Host:      host,
		Port:      port,
		User:      user,
		Password:  password,
		DBName:    dbName,
		DBTimeout: dbTimeout,
		DBDSN:     DBDSN,
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

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault("HOST", "localhost")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("GRPC_HOST", "0.0.0.0")
	viper.SetDefault("GRPC_PORT", "50051")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "postgres")
	viper.SetDefault("DB_TIMEOUT", "5s")

	viper.SetDefault("DEBUG", false)
	viper.SetDefault("LOGGING_MODE", "text")
	viper.SetDefault("JWT_SECRET_KEY", "dev-jwt-secret-key-change-in-prod!!")
	viper.SetDefault("ACCESS_TOKEN_EXPIRATION", 3600)
	viper.SetDefault("REFRESH_TOKEN_EXPIRATION", 86400)
	viper.SetDefault("TIMEOUT", "5s")
	viper.SetDefault("DB_SSL", "require")

	viper.SetDefault("THROTTLE_ENABLED", true)
	viper.SetDefault("THROTTLE_LIMIT", 300)
	viper.SetDefault("THROTTLE_WINDOW", "60s")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_PASSWORD", "123")
}
