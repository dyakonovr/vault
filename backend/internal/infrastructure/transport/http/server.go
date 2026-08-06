package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"time"
	"vault/internal/infrastructure/transport/http/auth"
	"vault/internal/infrastructure/transport/http/common"
	httpcommon "vault/internal/infrastructure/transport/http/common"
	"vault/internal/infrastructure/transport/http/wallet"

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

func (s *Server) InitRoutes(
	authMiddleware common.AuthMiddleware,
	authHandler *auth.AuthHandler,
	walletHandler *wallet.WalletHandler,
) {
	s.instance.GET("/health", func(c *echo.Context) error {
		return c.String(nethttp.StatusOK, "ok")
	})

	s.instance.Use(httpcommon.RequestIDMiddleware)

	// AUTH
	authGroup := s.instance.Group("/api/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/logout", authHandler.Logout)

		// защищённые auth-роуты
		authProtected := authGroup.Group("", authMiddleware.RequireAuth)
		{
			authProtected.GET("/me", authHandler.Me)
		}
	}

	// WALLETS
	wallets := s.instance.Group("/api/wallets", authMiddleware.RequireAuth)
	{
		wallets.POST("", walletHandler.Create)                    // POST   /api/wallets
		wallets.GET("/:id/balance", walletHandler.GetBalanceByID) // GET    /api/wallets/:id/balance
		wallets.POST("/:id/deposit", walletHandler.Deposit)       // POST   /api/wallets/:id/deposit
		wallets.POST("/:id/withdraw", walletHandler.Withdraw)     // POST   /api/wallets/:id/withdraw
	}
}
