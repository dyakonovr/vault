package wallet

type CreateWalletCommand struct {
	UserID  int64
	Balance int64
}

type UpdateWalletCommand struct {
	UserID int64
	Amount int64
}
