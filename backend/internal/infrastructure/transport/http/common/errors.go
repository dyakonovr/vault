package common

import (
	nethttp "net/http"
)

// ------------- GENERAL ERRORS -------------

var (
	ErrUnauthorized        = &HttpError{ClientMessage: "user unauthorized", Code: "UNAUTHORIZED", StatusCode: nethttp.StatusUnauthorized}
	ErrBadRequest          = &HttpError{ClientMessage: "bad request", Code: "BAD_REQUEST", StatusCode: nethttp.StatusBadRequest}
	ErrInternalServerError = &HttpError{ClientMessage: "internal server error", Code: "INTERNAL_SERVER_ERROR", StatusCode: nethttp.StatusInternalServerError}
	ErrUnprocessableEntity = &HttpError{ClientMessage: "unprocessable entity", Code: "UNPROCESSABLE_ENTITY", StatusCode: nethttp.StatusUnprocessableEntity}
	ErrValidation          = &HttpError{ClientMessage: "validation error", Code: "VALIDATION_ERROR", StatusCode: nethttp.StatusUnprocessableEntity}
)

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
	"issue_status":   "ISSUE_STATUS_INVALID",
	"issue_priority": "ISSUE_PRIORITY_INVALID",
}

// getValidationCode возвращает код ошибки по тегу.
// Если тег неизвестен, возвращается сам тег (fallback).
func getValidationCode(tag string) string {
	if code, ok := validationCodeMap[tag]; ok {
		return code
	}
	return tag
}

// ------------- DOMAIN ERROR MAPPING -------------

type DomainToHttpErrorMap map[error]error
func MapDomainErrorToHttp(domainErr error, domainToHttpMap DomainToHttpErrorMap) error {
	if domainErr == nil {
		return nil
	}

	if domainToHttpMap == nil {
		return domainErr
	}

	if e, ok := domainToHttpMap[domainErr]; ok {
		return e
	}

	return domainErr
}
