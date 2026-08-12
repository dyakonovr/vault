package common

import "github.com/labstack/echo/v5"

func ParseAndValidateRequestBody(ctx *echo.Context, req any) error {
	if err := ctx.Bind(req); err != nil {
		return err
	}
	if err := ctx.Validate(req); err != nil {
		return err
	}

	return nil
}

func ParseAndValidateQueryParams(ctx *echo.Context, req any) error {
	return ParseAndValidateRequestBody(ctx, req)
}

func GetStringPathParam(ctx *echo.Context, name string) string {
	return ctx.Param(name)
}

func GetIntPathParam(ctx *echo.Context, name string) (int64, error) {
	return echo.PathParam[int64](ctx, name)
}

func GetIDPathParam(ctx *echo.Context) (int64, error) {
	return GetIntPathParam(ctx, "id")
}