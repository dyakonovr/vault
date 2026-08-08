package common

import (
	"net/http"
	nethttp "net/http"
	"vault/internal/application/wallet"
	"vault/internal/domain"
)

// ------------- GENERAL ERRORS -------------

var (
	ErrUnauthorized        = &HttpError{ClientMessage: "user unauthorized", Code: "UNAUTHORIZED", StatusCode: nethttp.StatusUnauthorized}
	ErrBadRequest          = &HttpError{ClientMessage: "bad request", Code: "BAD_REQUEST", StatusCode: nethttp.StatusBadRequest}
	ErrInternalServerError = &HttpError{ClientMessage: "internal server error", Code: "INTERNAL_SERVER_ERROR", StatusCode: nethttp.StatusInternalServerError}
	ErrUnprocessableEntity = &HttpError{ClientMessage: "unprocessable entity", Code: "UNPROCESSABLE_ENTITY", StatusCode: nethttp.StatusUnprocessableEntity}
	ErrValidation          = &HttpError{ClientMessage: "validation error", Code: "VALIDATION_ERROR", StatusCode: nethttp.StatusUnprocessableEntity}
)

// ------------- DOMAIN TO HTTP ERRORS & MAPPING UTIL -------------

type DomainToHttpErrorMap map[error]error

func mapDomainErrorToHttp(domainErr error) error {
	if domainErr == nil {
		return nil
	}

	if domainErrorsToHttp == nil {
		return domainErr
	}

	if e, ok := domainErrorsToHttp[domainErr]; ok {
		return e
	}

	return domainErr
}

var domainErrorsToHttp = DomainToHttpErrorMap{
	// USER & AUTH
	domain.ErrUserNotFound:           &HttpError{ClientMessage: domain.ErrUserNotFound.Error(), Code: "USER_NOT_FOUND", StatusCode: http.StatusNotFound},
	domain.ErrUserAlreadyExists:      &HttpError{ClientMessage: domain.ErrUserAlreadyExists.Error(), Code: "USER_ALREADY_EXISTS", StatusCode: http.StatusConflict},
	domain.ErrInvalidLoginOrPassword: &HttpError{ClientMessage: domain.ErrInvalidLoginOrPassword.Error(), Code: "INVALID_LOGIN_OR_PASSWORD", StatusCode: http.StatusUnauthorized},
	domain.ErrWrongPassword:          &HttpError{ClientMessage: domain.ErrWrongPassword.Error(), Code: "ERROR_WRONG_PASSWORD", StatusCode: http.StatusForbidden},
	// TRANSACTIONS
	domain.ErrTransactionNotFound: &HttpError{ClientMessage: domain.ErrTransactionNotFound.Error(), Code: "TRANSACTION_NOT_FOUND", StatusCode: http.StatusNotFound},
	// WALLET
	domain.ErrWalletNotFound:      &HttpError{ClientMessage: domain.ErrWalletNotFound.Error(), Code: "WALLET_NOT_FOUND", StatusCode: http.StatusNotFound},
	domain.ErrWalletAlreadyExists: &HttpError{ClientMessage: domain.ErrWalletAlreadyExists.Error(), Code: "WALLET_ALREADY_EXISTS", StatusCode: http.StatusConflict},
	wallet.ErrWalletAccessDenied:  &HttpError{ClientMessage: wallet.ErrWalletAccessDenied.Error(), Code: "WALLET_ACCESS_DENIED", StatusCode: http.StatusForbidden},
	domain.ErrWalletInvalidAmount: &HttpError{ClientMessage: domain.ErrWalletInvalidAmount.Error(), Code: "WALLET_ACTION_INVALID_AMOUNT", StatusCode: http.StatusBadRequest},
	domain.ErrInsufficientFunds:   &HttpError{ClientMessage: domain.ErrInsufficientFunds.Error(), Code: "WALLET_INSUFFICIENT_FUNDS_ON_BALANCE", StatusCode: http.StatusConflict},
}

// ------------- VALIDATION -------------

// validationCodeMap сопоставляет тег валидатора (из fieldErr.Tag()) с внутренним кодом ошибки.
// Кастомные валидаторы добавляются сюда же по мере регистрации.
var validationCodeMap = map[string]string{
	"required": "FIELD_REQUIRED",
	"email":    "FIELD_INVALID_EMAIL",
	"url":      "FIELD_INVALID_URL",
	"uuid":     "FIELD_INVALID_UUID",
	"boolean":  "FIELD_INVALID_BOOLEAN",
	"numeric":  "FIELD_INVALID_NUMERIC",
	"min":      "FIELD_MIN_VALUE", // для чисел
	"max":      "FIELD_MAX_VALUE",
	"len":      "FIELD_LENGTH",
	"oneof":    "FIELD_NOT_ALLOWED_VALUE",
	"eq":       "FIELD_NOT_EQUAL",
	"ne":       "FIELD_NOT_EQUAL",
	"gt":       "FIELD_TOO_SMALL",
	"gte":      "FIELD_TOO_SMALL",
	"lt":       "FIELD_TOO_LARGE",
	"lte":      "FIELD_TOO_LARGE",
	// кастомные валидаторы:
	// "issue_status":   "ISSUE_STATUS_INVALID",
	// "issue_priority": "ISSUE_PRIORITY_INVALID",
}

// getValidationCode возвращает код ошибки по тегу.
// Если тег неизвестен, возвращается сам тег (fallback).
func getValidationCode(tag string) string {
	if code, ok := validationCodeMap[tag]; ok {
		return code
	}
	return tag
}
