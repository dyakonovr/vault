package auth

import (
	nethttp "net/http"
	authapp "vault/internal/application/auth"
	httpcommon "vault/internal/infrastructure/transport/http/common"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	authService authService
}

func New(authService authService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(ctx *echo.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	session, err := h.authService.Login(ctx.Request().Context(), authapp.LoginCommand{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, authDomainToHttpErrors))
	}

	h.setSessionCookie(ctx, session)

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusNoContent, nil)
}

func (h *AuthHandler) Register(ctx *echo.Context) error {
	var req RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	err := h.authService.Register(ctx.Request().Context(), authapp.RegisterCommand{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, authDomainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusNoContent, nil)
}

func (h *AuthHandler) Me(ctx *echo.Context) error {
	sessionCookie, err := httpcommon.ReadSessionID(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	user, err := h.authService.Me(ctx.Request().Context(), sessionCookie.Value)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, authDomainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusOK, NewUserResponse(user))
}

func (h *AuthHandler) Logout(ctx *echo.Context) error {
	sessionCookie, err := httpcommon.ReadSessionID(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	err = h.authService.Logout(ctx.Request().Context(), sessionCookie.Value)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, httpcommon.MapDomainErrorToHttp(err, authDomainToHttpErrors))
	}

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusNoContent, nil)
}

func (h *AuthHandler) setSessionCookie(ctx *echo.Context, session authapp.Session) {
	cookie := nethttp.Cookie{
		Name:     "session_id",
		Value:    session.Value,
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
		MaxAge:   int(session.Ttl.Seconds()),
	}

	ctx.SetCookie(&cookie)
}
