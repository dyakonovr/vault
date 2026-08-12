package wallet

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	depositapp "vault/internal/application/deposit"
	transferapp "vault/internal/application/transfer"
	walletapp "vault/internal/application/wallet"
	withdrawalapp "vault/internal/application/withdrawal"
	"vault/internal/domain"
	"vault/internal/infrastructure/transport/http/wallet/mocks"
	"vault/pkg/ctxkeys"

	"github.com/google/uuid"
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

func setupMiddleware(userID int64, idempotencyKey uuid.UUID) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := req.Context()
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, userID)
			ctx = context.WithValue(ctx, ctxkeys.RequestIDKey, "test-request-id")
			ctx = context.WithValue(ctx, ctxkeys.IdempotencyKey, idempotencyKey)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func TestCreate_Success(t *testing.T) {
	e := setupEcho()
	mockWalletService := mocks.NewMockWalletService(t)

	wallet := domain.Wallet{
		ID:        1,
		UserID:    10,
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockWalletService.On("Create", mock.Anything, walletapp.CreateWalletCommand{UserID: 10}).
		Return(wallet, nil)

	handler := New(mockWalletService, nil, nil, nil)
	e.POST("/api/wallets", handler.Create, setupMiddleware(10, uuid.Nil))

	req := httptest.NewRequest(http.MethodPost, "/api/wallets", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	mockWalletService.AssertExpectations(t)
}

func TestCreate_AlreadyExists(t *testing.T) {
	e := setupEcho()
	mockWalletService := mocks.NewMockWalletService(t)

	mockWalletService.On("Create", mock.Anything, walletapp.CreateWalletCommand{UserID: 10}).
		Return(domain.Wallet{}, domain.ErrWalletAlreadyExists)

	handler := New(mockWalletService, nil, nil, nil)
	e.POST("/api/wallets", handler.Create, setupMiddleware(10, uuid.Nil))

	req := httptest.NewRequest(http.MethodPost, "/api/wallets", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	mockWalletService.AssertExpectations(t)
}

func TestGetBalanceByID_Success(t *testing.T) {
	e := setupEcho()
	mockWalletService := mocks.NewMockWalletService(t)

	wallet := domain.Wallet{
		ID:        1,
		UserID:    10,
		Balance:   500,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockWalletService.On("GetById", mock.Anything, int64(1), int64(10)).
		Return(wallet, nil)

	handler := New(mockWalletService, nil, nil, nil)
	e.GET("/api/wallets/:id/balance", handler.GetBalanceByID, setupMiddleware(10, uuid.Nil))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1/balance", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"balance":500`)
	mockWalletService.AssertExpectations(t)
}

func TestGetBalanceByID_WalletNotFound(t *testing.T) {
	e := setupEcho()
	mockWalletService := mocks.NewMockWalletService(t)

	mockWalletService.On("GetById", mock.Anything, int64(999), int64(10)).
		Return(domain.Wallet{}, domain.ErrWalletNotFound)

	handler := New(mockWalletService, nil, nil, nil)
	e.GET("/api/wallets/:id/balance", handler.GetBalanceByID, setupMiddleware(10, uuid.Nil))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/999/balance", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	mockWalletService.AssertExpectations(t)
}

func TestGetBalanceByID_AccessDenied(t *testing.T) {
	e := setupEcho()
	mockWalletService := mocks.NewMockWalletService(t)

	mockWalletService.On("GetById", mock.Anything, int64(1), int64(10)).
		Return(domain.Wallet{}, walletapp.ErrWalletAccessDenied)

	handler := New(mockWalletService, nil, nil, nil)
	e.GET("/api/wallets/:id/balance", handler.GetBalanceByID, setupMiddleware(10, uuid.Nil))

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1/balance", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	mockWalletService.AssertExpectations(t)
}

func TestDeposit_Success(t *testing.T) {
	e := setupEcho()
	mockDepositService := mocks.NewMockDepositService(t)

	idempotencyKey := uuid.New()
	tx := domain.Transaction{
		ID:             1,
		WalletID:       1,
		Type:           domain.TransactionTypeDeposit,
		Amount:         100,
		IdempotencyKey: idempotencyKey,
		Status:         domain.TransactionStatusCompleted,
		CreatedAt:      time.Now(),
	}
	mockDepositService.On("Do", mock.Anything, depositapp.DepositCommand{
		UserID:         10,
		WalletID:       1,
		Amount:         100,
		IdempotencyKey: idempotencyKey,
	}).Return(tx, nil)

	handler := New(nil, mockDepositService, nil, nil)
	e.POST("/api/wallets/:id/deposit", handler.Deposit, setupMiddleware(10, idempotencyKey))

	body := `{"amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/wallets/1/deposit", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"amount":100`)
	require.Contains(t, rec.Body.String(), `"type":"deposit"`)
	mockDepositService.AssertExpectations(t)
}

func TestDeposit_InvalidAmount(t *testing.T) {
	e := setupEcho()
	mockDepositService := mocks.NewMockDepositService(t)

	handler := New(nil, mockDepositService, nil, nil)
	e.POST("/api/wallets/:id/deposit", handler.Deposit, setupMiddleware(10, uuid.New()))

	body := `{"amount":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/wallets/1/deposit", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestWithdrawal_Success(t *testing.T) {
	e := setupEcho()
	mockWithdrawalService := mocks.NewMockWithdrawalService(t)

	idempotencyKey := uuid.New()
	tx := domain.Transaction{
		ID:             1,
		WalletID:       1,
		Type:           domain.TransactionTypeWithdrawal,
		Amount:         50,
		IdempotencyKey: idempotencyKey,
		Status:         domain.TransactionStatusCompleted,
		CreatedAt:      time.Now(),
	}
	mockWithdrawalService.On("Do", mock.Anything, withdrawalapp.WithdrawalCommand{
		UserID:         10,
		WalletID:       1,
		Amount:         50,
		IdempotencyKey: idempotencyKey,
	}).Return(tx, nil)

	handler := New(nil, nil, mockWithdrawalService, nil)
	e.POST("/api/wallets/:id/withdrawal", handler.Withdrawal, setupMiddleware(10, idempotencyKey))

	body := `{"amount":50}`
	req := httptest.NewRequest(http.MethodPost, "/api/wallets/1/withdrawal", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"amount":50`)
	require.Contains(t, rec.Body.String(), `"type":"withdrawal"`)
	mockWithdrawalService.AssertExpectations(t)
}

func TestWithdrawal_InsufficientFunds(t *testing.T) {
	e := setupEcho()
	mockWithdrawalService := mocks.NewMockWithdrawalService(t)

	idempotencyKey := uuid.New()
	mockWithdrawalService.On("Do", mock.Anything, withdrawalapp.WithdrawalCommand{
		UserID:         10,
		WalletID:       1,
		Amount:         10000,
		IdempotencyKey: idempotencyKey,
	}).Return(domain.Transaction{}, domain.ErrInsufficientFunds)

	handler := New(nil, nil, mockWithdrawalService, nil)
	e.POST("/api/wallets/:id/withdrawal", handler.Withdrawal, setupMiddleware(10, idempotencyKey))

	body := `{"amount":10000}`
	req := httptest.NewRequest(http.MethodPost, "/api/wallets/1/withdrawal", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	mockWithdrawalService.AssertExpectations(t)
}

func TestTransfer_Success(t *testing.T) {
	e := setupEcho()
	mockTransferService := mocks.NewMockTransferService(t)

	idempotencyKey := uuid.New()
	tx := domain.Transaction{
		ID:             1,
		WalletID:       1,
		Type:           domain.TransactionTypeTransferOut,
		Amount:         200,
		IdempotencyKey: idempotencyKey,
		Status:         domain.TransactionStatusCompleted,
		CreatedAt:      time.Now(),
	}
	mockTransferService.On("Do", mock.Anything, transferapp.TransferCommand{
		UserID:         10,
		WalletFromID:   1,
		WalletToID:     2,
		Amount:         200,
		IdempotencyKey: idempotencyKey,
	}).Return(tx, nil)

	handler := New(nil, nil, nil, mockTransferService)
	e.POST("/api/wallets/transfers", handler.Transfer, setupMiddleware(10, idempotencyKey))

	body := `{"wallet_from_id":1,"wallet_to_id":2,"amount":200}`
	req := httptest.NewRequest(http.MethodPost, "/api/wallets/transfers", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"amount":200`)
	require.Contains(t, rec.Body.String(), `"type":"transfer_out"`)
	mockTransferService.AssertExpectations(t)
}

func TestTransfer_Validation(t *testing.T) {
	e := setupEcho()
	mockTransferService := mocks.NewMockTransferService(t)

	handler := New(nil, nil, nil, mockTransferService)
	e.POST("/api/wallets/transfers", handler.Transfer, setupMiddleware(10, uuid.New()))

	body := `{"wallet_from_id":0,"wallet_to_id":2,"amount":200}`
	req := httptest.NewRequest(http.MethodPost, "/api/wallets/transfers", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestGetBalanceByID_Unauthorized(t *testing.T) {
	e := setupEcho()
	mockWalletService := mocks.NewMockWalletService(t)

	handler := New(mockWalletService, nil, nil, nil)
	e.GET("/api/wallets/:id/balance", handler.GetBalanceByID, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := context.WithValue(req.Context(), ctxkeys.RequestIDKey, "test-request-id")
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1/balance", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
