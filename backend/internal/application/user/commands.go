package user

import "vault/internal/application"

type ListUsersCommand struct {
	Login *string
	application.PaginationCommand
}

type CreateUserCommand struct {
	Login    string
	Password string
}

type UpdateUserCommand struct {
	OldPassword string
	NewPassword string
}
