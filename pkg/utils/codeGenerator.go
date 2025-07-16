package utils

import (
	"crypto/rand"
)

type CodeGenerator struct{}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

func (cg *CodeGenerator) GenerateEmailVerificationCode(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)

	// Get all random bytes at once
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	for i := 0; i < length; i++ {
		b[i] = chars[b[i]%byte(len(chars))]
	}

	return string(b)
}
