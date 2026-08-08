package transaction

import httpcommon "vault/internal/infrastructure/transport/http/common"

type ListTransactionsFilters struct {
	Status *string `query:"status" validate:"omitempty,min=1"`
	Type   *string `query:"type"   validate:"omitempty,min=1"`
	httpcommon.PaginationParams
}
