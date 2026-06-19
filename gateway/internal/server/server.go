package server

import (
	"context"
	"errors"
	"gateway/internal/config"
	"gateway/internal/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
func NewServer(cfg *config.Config, h *handler.Handler, router *gin.Engine) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if h == nil {
		return nil, errors.New("handler is nil")
	}

	httpServer := &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.Timeout,
	}
	return &Server{
		cfg:        cfg,
		httpServer: httpServer,
	}, nil

}
