package main

import (
	"auth/internal/config"
	"auth/internal/grpcserver"
	"auth/internal/logging"
	"auth/internal/models"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/storage"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/subosito/gotenv"
	"go.uber.org/zap"
)

func main() {
	if err := gotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		return
	}
	config.Init()
	cfg := config.NewConfig()
	grpcCfg := config.NewGRPCConfig()
	throttleCfg := config.NewThrottleConfig()
	dbConfig := config.NewDBConfig()
	redisConfig := config.NewRedisConfig()
	ctx := context.Background()
	prefix := "refresh"

	if err := logging.InitLogger(cfg.LoggingMode); err != nil {
		fmt.Println(err)
		return
	}
	logger := logging.Logger

	logger.Info("Starting auth-service gRPC server",
		zap.String("host", grpcCfg.Host),
		zap.String("port", grpcCfg.Port),
	)

	db, err := storage.NewDatabase(dbConfig, models.ModelsList)
	if err != nil {
		logger.Fatal("Error while creating database", zap.Error(err))
	}

	redisClient := storage.NewRedisClient(redisConfig)
	accessTokenTTL := time.Duration(cfg.AccessTokenExpiration) * time.Second
	refreshTokenTTL := time.Duration(cfg.RefreshTokenExpiration) * time.Second

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(redisClient, ctx, prefix, cfg.JWTSecretKey, accessTokenTTL, "auth-service")

	userService := service.NewUserService(userRepo)
	tokenService := service.NewTokenService(tokenRepo, refreshTokenTTL)
	authService := service.NewAuthService(userService, tokenService)
	adminService := service.NewAdminService(userRepo)
	healthService := service.NewHealthService(userRepo, redisClient)

	grpcAuthServer := grpcserver.NewAuthServer(authService, adminService, healthService)
	grpcSrv, err := grpcserver.NewServer(grpcCfg, throttleCfg, grpcAuthServer)
	if err != nil {
		logger.Fatal("failed to create gRPC server", zap.Error(err))
	}

	go func() {
		if err := grpcSrv.Start(); err != nil {
			logger.Fatal("Error while starting gRPC server", zap.Error(err))
		}
	}()

	logger.Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	grpcSrv.Stop()
	logger.Info("Server stopped")
}
