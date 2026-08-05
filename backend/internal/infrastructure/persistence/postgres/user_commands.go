package postgres

type ListUsersParams struct {
	Login *string
	PaginationParams
}