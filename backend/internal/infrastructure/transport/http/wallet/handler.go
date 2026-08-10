package wallet

import (
	"net/http"
	depositapp "vault/internal/application/deposit"
	transferapp "vault/internal/application/transfer"
	walletapp "vault/internal/application/wallet"
	"vault/internal/application/withdrawal"
	httpcommon "vault/internal/infrastructure/transport/http/common"

	"github.com/labstack/echo/v5"
)

type WalletHandler struct {
	walletService     walletService
	depositService    depositService
	withdrawalService withdrawalService
	transferService   transferService
}

func New(walletService walletService, depositService depositService, withdrawalService withdrawalService, transferService transferService) *WalletHandler {
	return &WalletHandler{
		walletService:     walletService,
		depositService:    depositService,
		withdrawalService: withdrawalService,
		transferService:   transferService,
	}
}

// GetBalanceByID godoc
// @Summary      Получение баланса кошелька
// @Description  Возвращает текущий баланс кошелька по его идентификатору.
// @Tags         Кошельки
// @Produce      json
// @Security     session
// @Param        id   path     int true "ID кошелька"
// @Success      200  {object} BalanceResponse "Баланс кошелька"
// @Failure      400  {object} common.ErrorResponse "Некорректный идентификатор"
// @Failure      404  {object} common.ErrorResponse "Кошелёк не найден"
// @Failure      500  {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/wallets/{id}/balance [get]
func (h *WalletHandler) GetBalanceByID(ctx *echo.Context) error {
	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	wallet, err := h.walletService.GetById(ctx.Request().Context(), id, userID)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewBalanceResponse(wallet.Balance))
}

// Create godoc
// @Summary      Создание нового кошелька
// @Description  Создаёт новый кошелёк для текущего пользователя. Баланс нового кошелька равен 0.
// @Tags         Кошельки
// @Produce      json
// @Security     session
// @Success      201  {object} WalletResponse "Созданный кошелёк"
// @Failure      409  {object} common.ErrorResponse "Кошелёк уже существует"
// @Failure      500  {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/wallets [post]
func (h *WalletHandler) Create(ctx *echo.Context) error {
	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	wallet, err := h.walletService.Create(ctx.Request().Context(), walletapp.CreateWalletCommand{
		UserID: userID,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusCreated, NewWalletResponse(wallet))
}

// Deposit godoc
// @Summary      Пополнение кошелька
// @Description  Пополняет баланс кошелька на указанную сумму.
// @Tags         Кошельки
// @Accept       json
// @Produce      json
// @Security     session
// @Param        id   path     int            true "ID кошелька"
// @Param        body body     DepositRequest true  "Сумма пополнения"
// @Success      200  {object} BalanceResponse "Обновлённый баланс"
// @Failure      400  {object} common.ErrorResponse "Ошибка валидации"
// @Failure      404  {object} common.ErrorResponse "Кошелёк не найден"
// @Failure      409  {object} common.ErrorResponse "Кошелёк уже существует"
// @Failure      500  {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/wallets/{id}/deposit [post]
func (h *WalletHandler) Deposit(ctx *echo.Context) error {
	var req DepositRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	walletID, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	idempotencyKey, ok := httpcommon.GetIdempotencyKeyFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrIdempotencyKeyNotFound)
	}

	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	transaction, err := h.depositService.Do(ctx.Request().Context(), depositapp.DepositCommand{
		UserID:         userID,
		Amount:         req.Amount,
		WalletID:       walletID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	// TODO: подумать, что нужно вернуть
	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, nil)
}

// Withdrawal godoc
// @Summary      Списание средств с кошелька
// @Description  Списывает указанную сумму с баланса кошелька. Если средств недостаточно, возвращается ошибка.
// @Tags         Кошельки
// @Accept       json
// @Produce      json
// @Security     session
// @Param        id   path     int              true "ID кошелька"
// @Param        body body     WithdrawRequest  true  "Сумма списания"
// @Success      200  {object} BalanceResponse   "Обновлённый баланс"
// @Failure      400  {object} common.ErrorResponse   "Ошибка валидации"
// @Failure      404  {object} common.ErrorResponse   "Кошелёк не найден"
// @Failure      409  {object} common.ErrorResponse   "Недостаточно средств на балансе"
// @Failure      500  {object} common.ErrorResponse   "Внутренняя ошибка сервера"
// @Router       /api/wallets/{id}/withdrawal [post]
func (h *WalletHandler) Withdrawal(ctx *echo.Context) error {
	var req WithdrawRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	walletID, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	idempotencyKey, ok := httpcommon.GetIdempotencyKeyFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrIdempotencyKeyNotFound)
	}

	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	transaction, err := h.withdrawalService.Do(ctx.Request().Context(), withdrawal.WithdrawalCommand{
		UserID:         userID,
		Amount:         req.Amount,
		WalletID:       walletID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	// TODO: подумать, что нужно вернуть
	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, nil)
}

func (h *WalletHandler) Transfer(ctx *echo.Context) error {
	var req TransferRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	idempotencyKey, ok := httpcommon.GetIdempotencyKeyFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrIdempotencyKeyNotFound)
	}

	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	err := h.transferService.Do(ctx.Request().Context(), transferapp.TransferCommand{
		UserID:         userID,
		Amount:         req.Amount,
		WalletFromID:   req.WalletFromID,
		WalletToID:     req.WalletToID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	// TODO: подумать, что нужно вернуть
	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, nil)
}
