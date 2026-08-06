package wallet

import (
	"net/http"
	"vault/internal/application/wallet"
	httpcommon "vault/internal/infrastructure/transport/http/common"

	"github.com/labstack/echo/v5"
)

type WalletHandler struct {
	walletService walletService
}

func New(walletService walletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

func (h *WalletHandler) GetBalanceByID(ctx *echo.Context) error {
	// TODO: Middleware с проверкой на владельца
	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	wallet, err := h.walletService.GetById(ctx.Request().Context(), id)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, domainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewBalanceResponse(wallet.Balance))
}

func (h *WalletHandler) Create(ctx *echo.Context) error {
	// TODO: заменить на получение UserID из контекста
	userID, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	wallet, err := h.walletService.Create(ctx.Request().Context(), wallet.CreateWalletCommand{
		UserID: userID,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, domainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusCreated, NewWalletResponse(wallet))
}

func (h *WalletHandler) Deposit(ctx *echo.Context) error {
	var req DepositRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	// TODO: заменить на получение UserID из контекста
	userID, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	wallet, err := h.walletService.Deposit(ctx.Request().Context(), id, wallet.WalletDepositCommand{
		UserID: userID,
		Amount: req.Amount,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, domainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewBalanceResponse(wallet.Balance))
}

func (h *WalletHandler) Withdraw(ctx *echo.Context) error {
	var req WithdrawRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	// TODO: заменить на получение UserID из контекста
	userID, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	wallet, err := h.walletService.Withdraw(ctx.Request().Context(), id, wallet.WalletWithdrawCommand{
		UserID: userID,
		Amount: req.Amount,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, domainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewBalanceResponse(wallet.Balance))
}