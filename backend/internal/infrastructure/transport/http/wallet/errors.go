package wallet

import (
	"net/http"
	"vault/internal/application/wallet"
	"vault/internal/domain"
	"vault/internal/infrastructure/transport/http/common"
)

var domainToHttpErrors = common.DomainToHttpErrorMap{
	domain.ErrWalletNotFound:      &common.HttpError{ClientMessage: domain.ErrWalletNotFound.Error(), Code: "WALLET_NOT_FOUND", StatusCode: http.StatusNotFound},
	domain.ErrWalletAlreadyExists: &common.HttpError{ClientMessage: domain.ErrWalletAlreadyExists.Error(), Code: "WALLET_ALREADY_EXISTS", StatusCode: http.StatusConflict},
	wallet.ErrWalletAccessDenied:  &common.HttpError{ClientMessage: wallet.ErrWalletAccessDenied.Error(), Code: "WALLET_ACCESS_DENIED", StatusCode: http.StatusForbidden},
	domain.ErrWalletInvalidAmount: &common.HttpError{ClientMessage: domain.ErrWalletInvalidAmount.Error(), Code: "WALLET_ACTION_INVALID_AMOUNT", StatusCode: http.StatusBadRequest},
	domain.ErrInsufficientFunds:   &common.HttpError{ClientMessage: domain.ErrInsufficientFunds.Error(), Code: "WALLET_INSUFFICIENT_FUNDS_ON_BALANCE", StatusCode: http.StatusConflict},
}
