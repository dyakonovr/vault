package auth

import (
	"net/http"
	"vault/internal/domain"
	"vault/internal/infrastructure/transport/http/common"
)

var authDomainToHttpErrors = common.DomainToHttpErrorMap{
	domain.ErrUserNotFound:           &common.HttpError{ClientMessage: domain.ErrUserNotFound.Error(), Code: "USER_NOT_FOUND", StatusCode: http.StatusNotFound},
	domain.ErrUserAlreadyExists:      &common.HttpError{ClientMessage: domain.ErrUserAlreadyExists.Error(), Code: "USER_ALREADY_EXISTS", StatusCode: http.StatusConflict},
	domain.ErrInvalidLoginOrPassword: &common.HttpError{ClientMessage: domain.ErrInvalidLoginOrPassword.Error(), Code: "INVALID_LOGIN_OR_PASSWORD", StatusCode: http.StatusUnauthorized},
	domain.ErrWrongPassword:          &common.HttpError{ClientMessage: domain.ErrWrongPassword.Error(), Code: "ERROR_WRONG_PASSWORD", StatusCode: http.StatusForbidden},
}
