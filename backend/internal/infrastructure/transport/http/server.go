package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"time"
	"vault/internal/infrastructure/transport/http/auth"
	"vault/internal/infrastructure/transport/http/common"
	httpcommon "vault/internal/infrastructure/transport/http/common"
	"vault/internal/infrastructure/transport/http/transaction"
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
	authMiddleware *common.AuthMiddleware,
	walletOwnershipMiddleware *common.WalletOwnershipMiddleware,
	authHandler *auth.AuthHandler,
	walletHandler *wallet.WalletHandler,
	transactionHandler *transaction.TransactionHandler,
) {
	s.instance.GET("/health", func(c *echo.Context) error {
		return c.String(nethttp.StatusOK, "ok")
	})

	apiGroup := s.instance.Group("/api")
	{
		s.instance.Use(httpcommon.RequestIDMiddleware)

		// AUTH
		authGroup := apiGroup.Group("/auth")
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
		wallets := apiGroup.Group("/wallets", authMiddleware.RequireAuth)
		{
			wallets.POST("", walletHandler.Create) // POST   /api/wallets

			walletOwnership := wallets.Group("", walletOwnershipMiddleware.Check)
			{
				walletOwnership.GET("/:id/balance", walletHandler.GetBalanceByID) // GET    /api/wallets/:id/balance
				walletOwnership.POST("/:id/deposit", walletHandler.Deposit)       // POST   /api/wallets/:id/deposit
				walletOwnership.POST("/:id/withdrawal", walletHandler.Withdrawal) // POST   /api/wallets/:id/withdrawal

				{
					walletOwnership.GET("/:walletId/transactions", transactionHandler.List)        //  GET /api/wallets/:walletId/transactions
					walletOwnership.GET("/:walletId/transactions/:id", transactionHandler.GetByID) //  GET /api/wallets/:walletId/transactions/:id
				}
			}
		}
	}
}
