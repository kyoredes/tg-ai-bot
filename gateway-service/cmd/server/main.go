package main

import (
	"context"
	"errors"
	"fmt"
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
	"resty.dev/v3"
)

func main() {
	if err := gotenv.Load(".env"); err != nil {
		fmt.Println(err)
		return
	}
	config.Init()
	cfg := config.NewConfig()
	authConfig := config.NewAuthConfig()
	subConfig := config.NewSubConfig()
	// redisConfig := config.NewRedisConfig()
	devConfig := config.NewDevConfig()
	ctx := context.Background()
	restyClient := resty.New()

	// redisClient := storage.NewRedisClient(redisConfig)

	logging.InitLogger(cfg.LoggingMode)
	logger := logging.Logger

	logger.Info("Starting server... with", zap.String("host", cfg.Host), zap.String("port", cfg.Port))

	telegramService := service.NewTelegramService(authConfig, subConfig, restyClient)
	// serverTokenService, err := service.NewServerTokenService(ctx, cfg.ServerTokenTTL, redisClient)
	// if err != nil {
	// 	logger.Fatal("failed to create server token service", zap.Error(err))
	// }

	h := handler.NewHandler(telegramService)
	serverAuthMiddleware := middleware.DevAuthMiddleware(devConfig)
	router := router.SetupRouter(h, serverAuthMiddleware)

	srv, err := server.NewServer(cfg, h, router)
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
