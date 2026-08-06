package common

import (
	"math"
	"net/http"

	"github.com/labstack/echo/v5"
)

func ReadCookie(ctx *echo.Context, name string) (*http.Cookie, error) {
	return ctx.Cookie(name)
}

func ReadSessionID(ctx *echo.Context) (*http.Cookie, error) {
	cookie, err := ReadCookie(ctx, SessionCookieName)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return cookie, err
}

func CountTotalPages(total, perPage int) int {
	return int(math.Ceil(float64(total) / float64(perPage)))
}
