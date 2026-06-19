package main

import (
	"context"
	"errors"
	"fmt"
	"gateway/internal/client"
	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/logging"
	"gateway/internal/middleware"
	"gateway/internal/router"
	"gateway/internal/server"
	"gateway/internal/service"
	"net/http"
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
	authConfig := config.NewAuthConfig()
	subConfig := config.NewSubConfig()
	aiConfig := config.NewAIConfig()
	devConfig := config.NewDevConfig()
	adminConfig := config.NewAdminConfig()

	if err := logging.InitLogger(cfg.LoggingMode); err != nil {
		fmt.Println(err)
		return
	}
	logger := logging.Logger

	logger.Info("Starting server... with", zap.String("host", cfg.Host), zap.String("port", cfg.Port))

	grpcClients, err := client.NewClients(authConfig, subConfig, aiConfig)
	if err != nil {
		logger.Fatal("failed to create gRPC clients", zap.Error(err))
	}
	defer grpcClients.Close()

	telegramService := service.NewTelegramService(grpcClients, cfg.Timeout)
	adminService := service.NewAdminService(
		grpcClients,
		adminConfig,
		cfg.Timeout,
	)

	h := handler.NewHandler(telegramService, adminService)
	serverAuthMiddleware := middleware.DevAuthMiddleware(devConfig)
	adminAuthMiddleware := middleware.AdminAuthMiddleware(adminService)
	corsMiddleware := middleware.CORSMiddleware(adminConfig.CORSOrigin)
	router := router.SetupRouter(h, serverAuthMiddleware, adminAuthMiddleware, corsMiddleware)

	srv, err := server.NewServer(cfg, router)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Error while starting server", zap.Error(err))
		}
	}()

	logger.Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("Error while stopping server", zap.Error(err))
	}

	logger.Info("Server stopped")
}
