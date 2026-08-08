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

	wallet, err := h.walletService.Create(ctx.Request().Context(), wallet.CreateWalletCommand{
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

	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	wallet, err := h.walletService.Deposit(ctx.Request().Context(), id, wallet.WalletDepositCommand{
		UserID: userID,
		Amount: req.Amount,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewBalanceResponse(wallet.Balance))
}

// Withdraw godoc
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
// @Router       /api/wallets/{id}/withdraw [post]
func (h *WalletHandler) Withdraw(ctx *echo.Context) error {
	var req WithdrawRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	wallet, err := h.walletService.Withdraw(ctx.Request().Context(), id, wallet.WalletWithdrawCommand{
		UserID: userID,
		Amount: req.Amount,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewBalanceResponse(wallet.Balance))
}
