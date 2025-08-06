package jwt

import (
	"errors"
	"fmt"
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

var ErrTokenExpired = errors.New("access token expired")

func (s *Session) GenerateAccessToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email":      email,
		"exp":        time.Now().UTC().UTC().Add(time.Hour * 24).Unix(),
		"token_type": "access",
		"iat":        time.Now().UTC().UTC().Unix(),
		"iss":        "nobelium24",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *Session) GenerateRefreshToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email":      email,
		"exp":        time.Now().UTC().UTC().Add(time.Hour * 24 * 30).Unix(),
		"token_type": "refresh",
		"iat":        time.Now().UTC().UTC().Unix(),
		"iss":        "nobelium24",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *Session) VerifyAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})

	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if ok && ve.Errors&jwt.ValidationErrorExpired != 0 {
			return "", ErrTokenExpired
		}
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email := claims["sub"].(string)
		return email, nil
	}

	return "", fmt.Errorf("invalid token")
}

func (s *Session) VerifyExpiredToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})
	if err != nil {
		return false, fmt.Errorf("error parsing token: %s", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		expUnix, ok := claims["exp"].(float64)
		if !ok {
			return false, fmt.Errorf("invalid or missing 'exp' field in token")
		}

		expTime := time.Unix(int64(expUnix), 0)
		if time.Now().After(expTime) {
			return true, nil
		}
		return false, nil
	}

	return false, fmt.Errorf("invalid token claims")
}

func (s *Session) VerifyRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})
	if err != nil {
		return "", fmt.Errorf("error parsing token: %s", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "refresh" {
			return "", fmt.Errorf("invalid token type: expected refresh token")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims: email not found")
		}
		accessToken, err := s.GenerateAccessToken(email)
		if err != nil {
			return "", err
		}
		return accessToken, nil
	}
	return "", fmt.Errorf("invalid token")
}
