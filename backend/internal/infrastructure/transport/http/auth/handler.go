package auth

import (
	nethttp "net/http"
	authapp "vault/internal/application/auth"
	httpcommon "vault/internal/infrastructure/transport/http/common"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	authService AuthService
}

func New(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по логину и паролю. При успешном входе устанавливается сессионная cookie.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        body body     LoginRequest true  "Данные для входа"
// @Success      204  {string} string         "Вход выполнен успешно"
// @Failure      400  {object} common.ErrorResponse "Ошибка валидации"
// @Failure      401  {object} common.ErrorResponse "Неверный логин или пароль"
// @Failure      500  {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(ctx *echo.Context) error {
	var req LoginRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	session, err := h.authService.Login(ctx.Request().Context(), authapp.LoginCommand{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	h.setSessionCookie(ctx, session)

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusNoContent, nil)
}

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создание новой учётной записи с логином и паролем.
// @Tags         Аутентификация
// @Accept       json
// @Produce      json
// @Param        body body     RegisterRequest true  "Данные для регистрации"
// @Success      204  {string} string          "Регистрация выполнена успешно"
// @Failure      400  {object} common.ErrorResponse "Ошибка валидации"
// @Failure      409  {object} common.ErrorResponse "Пользователь уже существует"
// @Failure      500  {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(ctx *echo.Context) error {
	var req RegisterRequest
	if err := httpcommon.ParseAndValidateRequestBody(ctx, &req); err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	err := h.authService.Register(ctx.Request().Context(), authapp.RegisterCommand{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusNoContent, nil)
}

// Me godoc
// @Summary      Получение данных текущего пользователя
// @Description  Возвращает информацию о пользователе по сессионной cookie.
// @Tags         Аутентификация
// @Produce      json
// @Security     session
// @Success      200 {object} UserResponse    "Данные пользователя"
// @Failure      401 {object} common.ErrorResponse "Сессия не найдена или истекла"
// @Failure      404 {object} common.ErrorResponse "Пользователь не найден"
// @Failure      500 {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/auth/me [get]
func (h *AuthHandler) Me(ctx *echo.Context) error {
	sessionCookie, err := httpcommon.ReadSessionID(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	user, err := h.authService.Me(ctx.Request().Context(), sessionCookie.Value)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusOK, NewUserResponse(user))
}

// Logout godoc
// @Summary      Выход из системы
// @Description  Удаление текущей сессии. Сессионная cookie аннулируется.
// @Tags         Аутентификация
// @Produce      json
// @Security     session
// @Success      204 {string} string         "Выход выполнен успешно"
// @Failure      401 {object} common.ErrorResponse "Сессия не найдена"
// @Failure      500 {object} common.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(ctx *echo.Context) error {
	sessionCookie, err := httpcommon.ReadSessionID(ctx)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	err = h.authService.Logout(ctx.Request().Context(), sessionCookie.Value)
	if err != nil {
		return httpcommon.HTTPErrorResponse(ctx, err)
	}

	return httpcommon.HTTPSuccessResponse(ctx, nethttp.StatusNoContent, nil)
}

func (h *AuthHandler) setSessionCookie(ctx *echo.Context, session authapp.Session) {
	cookie := nethttp.Cookie{
		Name:     httpcommon.SessionCookieName,
		Value:    session.Value,
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
		MaxAge:   int(session.Ttl.Seconds()),
	}

	ctx.SetCookie(&cookie)
}
