package common

import (
	"context"
	"vault/pkg/ctxkeys"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// ---------- Utils ----------

func setValueIntoRequestContext(c *echo.Context, key, val any) {
	req := c.Request()
	ctx := context.WithValue(req.Context(), key, val)
	c.SetRequest(req.WithContext(ctx))
}

// ---------- AUTH ----------

type sessionStore interface {
	GetUserID(ctx context.Context, sessionID string) (int64, error)
}

type AuthMiddleware struct {
	sessionStore sessionStore
}

func NewAuthMiddleware(sessionStore sessionStore) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore: sessionStore,
	}
}

func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		sessionCookie, err := c.Cookie(SessionCookieName)
		if err != nil {
			return HTTPErrorResponse(c, ErrUnauthorized)
		}

		userID, err := m.sessionStore.GetUserID(c.Request().Context(), sessionCookie.Value)
		if err != nil {
			return HTTPErrorResponse(c, ErrUnauthorized)
		}

		setValueIntoRequestContext(c, ctxkeys.UserIDKey, userID)
		return next(c)
	}
}

// ---------- REQUEST ID ----------

// RequestIDMiddleware генерирует уникальный идентификатор запроса,
// помещает его в контекст и в заголовок ответа X-Request-ID.
func RequestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		id := uuid.New().String()
		c.Response().Header().Set("X-Request-ID", id)
		setValueIntoRequestContext(c, ctxkeys.RequestIDKey, id)

		return next(c)
	}
}
