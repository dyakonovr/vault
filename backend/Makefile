# make migrate_create NAME=<name>
migrate_create:
	migrate create -ext sql -dir migrations -seq $(NAME)

swagger:
	swag init -g cmd/app/main.go