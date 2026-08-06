package wallet

type CreateWalletCommand struct {
	UserID int64
}

type WalletDepositCommand struct {
	UserID int64
	Amount int64
}

type WalletWithdrawCommand struct {
	UserID int64
	Amount int64
}
