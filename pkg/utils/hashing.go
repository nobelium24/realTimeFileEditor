package utils

import (
	"encoding/base64"
	"fmt"
	"realTimeEditor/pkg/constants"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher struct {
	Salt []byte
}

func NewPasswordHasher() (*PasswordHasher, error) {
	env, err := constants.LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("invalid salt provided")
	}
	return &PasswordHasher{Salt: []byte(env.SALT)}, nil
}

func (p *PasswordHasher) HashPassword(password string) (string, error) {
	hash := argon2.IDKey([]byte(password), p.Salt, 1, 64*1024, 4, 32)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	return encodedHash, nil
}

func (p *PasswordHasher) VerifyPassword(hashedPassword, password string) (bool, error) {

	salt := []byte(p.Salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	return encodedHash == hashedPassword, nil
}
