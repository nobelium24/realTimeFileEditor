package jwt

import (
	"realTimeEditor/pkg/constants"
	"time"

	"github.com/golang-jwt/jwt"
)

type Session struct {
	JWTSecret string
}

func NewSession() (*Session, error) {
	env, err := constants.LoadEnv()
	if err != nil {
		return nil, err
	}
	return &Session{
		JWTSecret: env.JWT_SECRET,
	}, nil
}

func (s *Session) GenerateAccessToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email":      email,
		"exp":        time.Now().UTC().Add(time.Hour * 24).Unix(),
		"token_type": "access",
		"iat":        time.Now().UTC().Unix(),
		"iss":        "nobelium24",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *Session) GenerateRefreshToken()
