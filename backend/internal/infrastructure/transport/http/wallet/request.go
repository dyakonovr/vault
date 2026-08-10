package wallet

type DepositRequest struct {
	Amount int64 `json:"amount"          validate:"required,min=1"`
}

type WithdrawRequest struct {
	Amount int64 `json:"amount"          validate:"required,min=1"`
}

type TransferRequest struct {
	WalletFromID int64 `json:"wallet_from_id" validate:"required,min=1"`
	WalletToID   int64 `json:"wallet_to_id" validate:"required,min=1"`
	Amount       int64 `json:"amount"          validate:"required,min=1"`
}
