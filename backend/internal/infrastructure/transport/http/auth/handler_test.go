package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	authapp "vault/internal/application/auth"
	"vault/internal/domain"
	"vault/internal/infrastructure/transport/http/auth/mocks"
	httpcommon "vault/internal/infrastructure/transport/http/common"
	"vault/pkg/ctxkeys"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type customValidator struct {
	v *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.v.Struct(i)
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &customValidator{v: validator.New()}
	return e
}

func setupContextWithUserID(userID int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := context.WithValue(req.Context(), ctxkeys.UserIDKey, userID)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func setupContextWithRequestID(requestID string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := context.WithValue(req.Context(), ctxkeys.RequestIDKey, requestID)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func TestLogin_Success(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	session := authapp.Session{
		Value: "session-token-123",
		Ttl:   1 * time.Hour,
	}
	mockService.On("Login", mock.Anything, mock.Anything).
		Return(session, nil)

	handler := New(mockService)

	e.POST("/api/auth/login", handler.Login, setupContextWithRequestID("test-request-id"))

	body := `{"login":"alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	cookies := rec.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == httpcommon.SessionCookieName {
			found = true
			require.Equal(t, "session-token-123", cookie.Value)
		}
	}
	require.True(t, found, "session cookie should be set")

	mockService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	mockService.On("Login", mock.Anything, mock.Anything).
		Return(authapp.Session{}, domain.ErrInvalidLoginOrPassword)

	handler := New(mockService)

	e.POST("/api/auth/login", handler.Login, setupContextWithRequestID("test-request-id"))

	body := `{"login":"alice","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)

	mockService.AssertExpectations(t)
}

func TestLogin_ValidationError(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	handler := New(mockService)

	e.POST("/api/auth/login", handler.Login, setupContextWithRequestID("test-request-id"))

	body := `{"login":"al","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestRegister_Success(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	mockService.On("Register", mock.Anything, mock.Anything).
		Return(nil)

	handler := New(mockService)

	e.POST("/api/auth/register", handler.Register, setupContextWithRequestID("test-request-id"))

	body := `{"login":"alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	mockService.AssertExpectations(t)
}

func TestRegister_AlreadyExists(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	mockService.On("Register", mock.Anything, mock.Anything).
		Return(domain.ErrUserAlreadyExists)

	handler := New(mockService)

	e.POST("/api/auth/register", handler.Register, setupContextWithRequestID("test-request-id"))

	body := `{"login":"alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestMe_Success(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	user := domain.User{
		ID:        1,
		Login:     "alice",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockService.On("Me", mock.Anything, "session-token-123").
		Return(user, nil)

	handler := New(mockService)

	e.GET("/api/auth/me", handler.Me,
		setupContextWithRequestID("test-request-id"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: httpcommon.SessionCookieName, Value: "session-token-123"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

func TestMe_NoSession(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	handler := New(mockService)

	e.GET("/api/auth/me", handler.Me,
		setupContextWithRequestID("test-request-id"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogout_Success(t *testing.T) {
	e := setupEcho()
	mockService := mocks.NewMockAuthService(t)

	mockService.On("Logout", mock.Anything, "session-token-123").
		Return(nil)

	handler := New(mockService)

	e.POST("/api/auth/logout", handler.Logout,
		setupContextWithRequestID("test-request-id"),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: httpcommon.SessionCookieName, Value: "session-token-123"})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	mockService.AssertExpectations(t)
}
