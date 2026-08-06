package common

import (
	"errors"
	nethttp "net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

func HTTPSuccessResponse(ctx *echo.Context, statusCode int, body any) error {
	if statusCode == nethttp.StatusNoContent {
		return ctx.NoContent(statusCode)
	}

	return ctx.JSON(statusCode, SuccessResponse{
		Data:      body,
		RequestID: "request_id",
	})
}

func HTTPErrorResponse(ctx *echo.Context, err error) error {
	var httpError *HttpError
	var validationError validator.ValidationErrors

	statusCode := ErrInternalServerError.StatusCode
	data := ErrorResponse{
		Code:    ErrInternalServerError.Code,
		Message: ErrInternalServerError.ClientMessage,
	}

	if errors.As(err, &httpError) {
		statusCode = httpError.StatusCode
		data = ErrorResponse{
			Code:    httpError.Code,
			Message: httpError.ClientMessage,
		}
	} else if errors.As(err, &validationError) {
		statusCode = nethttp.StatusUnprocessableEntity
		data = ErrorResponse{
			Code: ErrValidation.Code,
			Message: ErrValidation.ClientMessage,
			Errors: mapValidationErrorsIntoResponse(validationError),
		}
	} else {
		// Внутренняя ошибка: логируем с деталями
		// logger.FromContext(r.Context()).WithError(err).Error("internal server error")
	}

	data.RequestID = "request_id"

	return ctx.JSON(statusCode, data)
}
