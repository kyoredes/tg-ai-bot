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

func Init() {
	viper.AutomaticEnv()
	viper.SetDefault("HOST", "localhost")
	viper.SetDefault("PORT", "8099")
	viper.SetDefault("GRPC_HOST", "0.0.0.0")
	viper.SetDefault("GRPC_PORT", "50052")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5430)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "postgres")
	viper.SetDefault("DB_TIMEOUT", "5s")

	viper.SetDefault("DEBUG", false)
	viper.SetDefault("LOGGING_MODE", "text")
	viper.SetDefault("JWT_SECRET_KEY", "secret")
	viper.SetDefault("ACCESS_TOKEN_EXPIRATION", 3600)
	viper.SetDefault("REFRESH_TOKEN_EXPIRATION", 86400)
	viper.SetDefault("TIMEOUT", "5s")
	viper.SetDefault("DB_SSL", "require")
}
