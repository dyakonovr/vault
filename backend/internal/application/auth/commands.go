package auth

import "time"

type LoginCommand struct {
	Login    string
	Password string
}

type RegisterCommand struct {
	Login    string
	Password string
}

type Session struct {
	Value string
	Ttl   time.Duration
}
