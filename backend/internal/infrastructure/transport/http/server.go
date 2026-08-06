package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"time"
	"vault/internal/infrastructure/transport/http/auth"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type Server struct {
	instance *echo.Echo
	port     int
}

func NewServer(port int) *Server {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	return &Server{
		instance: e,
		port:     port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	sc := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", s.port),
		GracefulTimeout: 10 * time.Second, // время на завершение текущих запросов
	}
	return sc.Start(ctx, s.instance)
}

func (s *Server) InitRoutes(authHandler *auth.AuthHandler) {
	s.instance.GET("/health", func(c *echo.Context) error {
		return c.String(nethttp.StatusOK, "ok")
	})

	authGroup := s.instance.Group("/auth", nil)

	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.GET("/me", authHandler.Me)
}
