package transaction

import (
	"net/http"
	"vault/internal/application"
	"vault/internal/application/transaction"
	httpcommon "vault/internal/infrastructure/transport/http/common"

	"github.com/labstack/echo/v5"
)

type TransactionHandler struct {
	transactionService TransactionService
}

func New(transactionService TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// List godoc
// @Summary      Список транзакций кошелька
// @Description  Возвращает список транзакций кошелька с фильтрацией и пагинацией.
// @Tags         Транзакции
// @Produce      json
// @Security     session
// @Param        walletId path     int                    true "ID кошелька"
// @Param        body     body     ListTransactionsFilters false "Фильтры и параметры пагинации"
// @Success      200      {array}  TransactionResponse    "Список транзакций"
// @Failure      400      {object} common.ErrorResponse   "Некорректный идентификатор"
// @Failure      401      {object} common.ErrorResponse   "Необходима авторизация"
// @Failure      403      {object} common.ErrorResponse   "Доступ запрещён"
// @Failure      500      {object} common.ErrorResponse   "Внутренняя ошибка сервера"
// @Router       /api/wallets/{walletId}/transactions [get]
func (h *TransactionHandler) List(ctx *echo.Context) error {
	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	walletID, err := httpcommon.GetIntPathParam(ctx, "walletId")
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	filters := ListTransactionsFilters{
		PaginationParams: httpcommon.PaginationParams{
			Page:    1,
			PerPage: 20,
		},
	}
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &filters); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	transactions, total, err := h.transactionService.List(ctx.Request().Context(), transaction.ListTransactionsCommand{
		WalletID: walletID,
		UserID:   userID,
		Status:   filters.Status,
		Type:     filters.Type,
		PaginationCommand: application.PaginationCommand{
			Limit:  filters.PerPage,
			Offset: httpcommon.CountOffset(filters.Page, filters.PerPage),
		},
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	list := make([]TransactionResponse, len(transactions))
	for _, t := range transactions {
		list = append(list, NewTransactionResponse(t))
	}

	return httpcommon.HTTPPaginatedResponse(ctx, http.StatusOK, list, httpcommon.PaginatedResponseMeta{
		Page:       filters.Page,
		PerPage:    filters.PerPage,
		Total:      total,
		TotalPages: httpcommon.CountTotalPages(total, filters.PerPage),
	})
}

// GetByID godoc
// @Summary      Получение транзакции по ID
// @Description  Возвращает конкретную транзакцию по её идентификатору.
// @Tags         Транзакции
// @Produce      json
// @Security     session
// @Param        walletId  path     int                      true "ID кошелька"
// @Param        id        path     int                      true "ID транзакции"
// @Success      200       {object} TransactionResponse      "Данные транзакции"
// @Failure      400       {object} common.ErrorResponse     "Некорректный идентификатор"
// @Failure      401       {object} common.ErrorResponse     "Необходима авторизация"
// @Failure      403       {object} common.ErrorResponse     "Доступ запрещён"
// @Failure      404       {object} common.ErrorResponse     "Транзакция не найдена"
// @Failure      500       {object} common.ErrorResponse     "Внутренняя ошибка сервера"
// @Router       /api/wallets/{walletId}/transactions/{id} [get]
func (h *TransactionHandler) GetByID(ctx *echo.Context) error {
	userID, ok := httpcommon.GetUserIDFromContext(ctx.Request().Context())
	if !ok {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.ErrUnauthorized)
	}

	walletID, err := httpcommon.GetIntPathParam(ctx, "walletId")
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	id, err := httpcommon.GetIDPathParam(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	transaction, err := h.transactionService.GetById(ctx.Request().Context(), id, transaction.GetTransactionByIdCommand{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewTransactionResponse(transaction))
}
