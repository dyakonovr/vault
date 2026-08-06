package common

import (
	"context"
	"vault/pkg/ctxkeys"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// ---------- REQUEST ID ----------

// RequestIDMiddleware генерирует уникальный идентификатор запроса,
// помещает его в контекст и в заголовок ответа X-Request-ID.
func RequestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		id := uuid.New().String()
		c.Response().Header().Set("X-Request-ID", id)

		req := c.Request()
		ctx := context.WithValue(req.Context(), ctxkeys.RequestIDKey, id)
		c.SetRequest(req.WithContext(ctx))

		return next(c)
	}
}
