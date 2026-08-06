package main

// @title           Vault API
// @version         1.0
// @description     API для управления кошельками и аутентификации пользователей.
// @host            localhost:8080
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey session
// @in cookie
// @name session_id
import (
	"vault/internal/app"
)

func main() {
	app.Run()
}
