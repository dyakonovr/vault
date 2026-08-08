package transaction

import (
	"net/http"
	"vault/internal/application"
	"vault/internal/application/transaction"
	httpcommon "vault/internal/infrastructure/transport/http/common"

	"github.com/labstack/echo/v5"
)

type TransactionHandler struct {
	transactionService transactionService
}

func New(transactionService transactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

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
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, domainToHttpErrors))
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
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, domainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, http.StatusOK, NewTransactionResponse(transaction))
}
