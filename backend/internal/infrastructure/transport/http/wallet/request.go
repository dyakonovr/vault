package wallet

type DepositRequest struct {
	Amount int64 `json:"amount" validate:"required,min=1"`
}

type WithdrawRequest struct {
	Amount int64 `json:"amount" validate:"required,min=1"`
}
