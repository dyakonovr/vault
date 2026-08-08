package transaction

import (
	"net/http"
	"vault/internal/domain"
	"vault/internal/infrastructure/transport/http/common"
)

// TODO: Сейчас при ownership-проверке вылетит ошибка из другого домена,
// здесь она не перемаппиться, из-за чего будет 500?

var domainToHttpErrors = common.DomainToHttpErrorMap{
	domain.ErrTransactionNotFound: &common.HttpError{ClientMessage: domain.ErrTransactionNotFound.Error(), Code: "TRANSACTION_NOT_FOUND", StatusCode: http.StatusNotFound},
}
