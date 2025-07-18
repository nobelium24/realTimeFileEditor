package jwt

import (
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

func (s *Session) GenerateRefreshToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email":      email,
		"exp":        time.Now().UTC().Add(time.Hour * 24 * 30).Unix(),
		"token_type": "refresh",
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

func (s *Session) VerifyAccessToken(tokenString string) (string, error) {
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
		// Check token type
		if tokenType, ok := claims["role"].(string); !ok || tokenType != "member" {
			return "", fmt.Errorf("invalid token: expecting member token")
		}

		if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "access" {
			return "", fmt.Errorf("invalid token type: expected access token")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims: email not found")
		}
		return email, nil
	}
	return "", fmt.Errorf("invalid token")
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
		// Check token type
		if tokenType, ok := claims["role"].(string); !ok || tokenType != "member" {
			return "", fmt.Errorf("invalid token: expecting member token")
		}

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
