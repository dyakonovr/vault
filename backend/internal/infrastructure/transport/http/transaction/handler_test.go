package transaction

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	transactionapp "vault/internal/application/transaction"
	"vault/internal/domain"
	"vault/internal/infrastructure/transport/http/transaction/mocks"
	"vault/pkg/ctxkeys"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type customValidator struct {
	v *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.v.Struct(i)
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &customValidator{v: validator.New()}
	return e
}

func setupMiddleware(userID int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := req.Context()
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, userID)
			ctx = context.WithValue(ctx, ctxkeys.RequestIDKey, "test-request-id")
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func TestList_Success(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	now := time.Now()
	transactions := []domain.Transaction{
		{
			ID:        1,
			WalletID:  10,
			Type:      domain.TransactionTypeDeposit,
			Amount:    100,
			Status:    domain.TransactionStatusCompleted,
			CreatedAt: now,
		},
		{
			ID:        2,
			WalletID:  10,
			Type:      domain.TransactionTypeWithdrawal,
			Amount:    50,
			Status:    domain.TransactionStatusCompleted,
			CreatedAt: now,
		},
	}
	mockService.On("List", mock.Anything, mock.Anything).
		Return(transactions, int64(2), nil)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions", handler.List, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"deposit"`)
	require.Contains(t, rec.Body.String(), `"withdrawal"`)
	require.Contains(t, rec.Body.String(), `"total":2`)
	mockService.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	mockService.On("List", mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, int64(0), nil)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions", handler.List, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"total":0`)
	mockService.AssertExpectations(t)
}

func TestList_WithFilters(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	transactions := []domain.Transaction{
		{
			ID:        1,
			WalletID:  10,
			Type:      domain.TransactionTypeDeposit,
			Amount:    200,
			Status:    domain.TransactionStatusCompleted,
			CreatedAt: time.Now(),
		},
	}
	mockService.On("List", mock.Anything, mock.Anything).
		Return(transactions, int64(1), nil)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions", handler.List, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions?status=completed&type=deposit", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"deposit"`)
	mockService.AssertExpectations(t)
}

func TestGetByID_Success(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	tx := domain.Transaction{
		ID:        1,
		WalletID:  10,
		Type:      domain.TransactionTypeDeposit,
		Amount:    100,
		Status:    domain.TransactionStatusCompleted,
		CreatedAt: time.Now(),
	}
	mockService.On("GetById", mock.Anything, int64(1), mock.Anything).
		Return(tx, nil)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions/:id", handler.GetByID, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions/1", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"id":1`)
	require.Contains(t, rec.Body.String(), `"deposit"`)
	require.Contains(t, rec.Body.String(), `"amount":100`)
	mockService.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	mockService.On("GetById", mock.Anything, int64(999), mock.Anything).
		Return(domain.Transaction{}, domain.ErrTransactionNotFound)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions/:id", handler.GetByID, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions/999", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}

func TestGetByID_InvalidID(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions/:id", handler.GetByID, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/abc/transactions/1", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, rec.Code)
}

func TestList_Unauthorized(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions", handler.List, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := context.WithValue(req.Context(), ctxkeys.RequestIDKey, "test-request-id")
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestList_Pagination(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockTransactionService(t)

	transactions := []domain.Transaction{
		{ID: 3, WalletID: 10, Type: domain.TransactionTypeDeposit, Amount: 50, Status: domain.TransactionStatusCompleted, CreatedAt: time.Now()},
	}
	mockService.On("List", mock.Anything, mock.MatchedBy(func(cmd transactionapp.ListTransactionsCommand) bool {
		return cmd.WalletID == 10 && cmd.Offset == 20 && cmd.Limit == 10
	})).Return(transactions, int64(30), nil)

	handler := New(mockService)
	e.GET("/api/wallets/:walletId/transactions", handler.List, setupMiddleware(10))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/10/transactions?page=3&per_page=10", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"total":30`)
	require.Contains(t, rec.Body.String(), `"page":3`)
	require.Contains(t, rec.Body.String(), `"per_page":10`)
	mockService.AssertExpectations(t)
}
