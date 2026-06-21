package main

import (
	"fmt"
	"os"
	"os/signal"
	"subscription/internal/config"
	"subscription/internal/grpcserver"
	"subscription/internal/logging"
	"subscription/internal/models"
	"subscription/internal/repository"
	"subscription/internal/service"
	"subscription/internal/storage"
	"syscall"

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

	if err := logging.InitLogger(cfg.LoggingMode); err != nil {
		fmt.Println(err)
		return
	}
	logger := logging.Logger

	logger.Info("Starting subscription-service gRPC server",
		zap.String("host", grpcCfg.Host),
		zap.String("port", grpcCfg.Port),
	)

	dbConfig := config.NewDBConfig()
	db, err := storage.NewDatabase(dbConfig, models.ModelsList)
	if err != nil {
		logger.Fatal("failed to create database", zap.Error(err))
	}

	subscriptionRepo := repository.NewSubscriptionRepository(db)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	adminService := service.NewAdminService(subscriptionRepo)
	healthService := service.NewHealthService(subscriptionRepo)

	grpcSubServer := grpcserver.NewSubscriptionServer(subscriptionService, adminService, healthService)
	grpcSrv, err := grpcserver.NewServer(grpcCfg, throttleCfg, grpcSubServer)
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
