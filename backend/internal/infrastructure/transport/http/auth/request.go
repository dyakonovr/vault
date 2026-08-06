package auth

type LoginRequest struct {
	Login    string `json:"login" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Login    string `json:"login" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
}
