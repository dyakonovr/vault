package hash

import "github.com/alexedwards/argon2id"

func HashArgon2(value string) (string, error) {
	return argon2id.CreateHash(value, argon2id.DefaultParams)
}

func CompareArgon2(password, passwordHash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, passwordHash)
}
