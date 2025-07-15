package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"realTimeEditor/internal/handlers"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
	"realTimeEditor/pkg/jwt"
	"realTimeEditor/pkg/utils"
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	UserRepository           repositories.UserRepository
	ForgotPasswordRepository repositories.ForgotPasswordRepository
}

func NewUserHandler(
	userRepository *repositories.UserRepository,
	forgotPasswordRepository *repositories.ForgotPasswordRepository,
) *UserHandler {
	return &UserHandler{
		UserRepository:           *userRepository,
		ForgotPasswordRepository: *forgotPasswordRepository,
	}
}

var (
	emailRegex    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	nameRegex     = regexp2.MustCompile(`^[\p{L}\p{M}\p{Zs}\-'.]+$`, 0)
	passwordRegex = regexp.MustCompile(`^[\x20-\x7E]{6,}$`)
	codeRegex     = regexp.MustCompile(`^[A-Za-z0-9]{6}$`)
)

func (u *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Error binding JSON: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	var existingUser model.User

	ok, err := nameRegex.MatchString(*user.FirstName)
	if err != nil {
		log.Printf("Regex match error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid first name format"})
		return
	}

	ok, err = nameRegex.MatchString(*user.LastName)
	if err != nil {
		log.Printf("Regex match error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last name format"})
		return
	}

	if !emailRegex.MatchString(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	if !passwordRegex.MatchString(*user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password format"})
		return
	}

	err = u.UserRepository.GetByEmail(&existingUser, user.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error retrieving user details: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if existingUser.Email != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	err = u.UserRepository.GetByEmail(&existingUser, user.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error retrieving user details: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	passwordHasher, err := utils.NewPasswordHasher()
	if err != nil {
		log.Printf("Error creating password hasher: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	hashedPassword, err := passwordHasher.HashPassword(*user.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	user.Password = &hashedPassword

	_, err = u.UserRepository.Create(&user)

	if err != nil {
		log.Printf("Error creating user: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	err = handlers.SendMail(user.Email, "welcome", "Welcome Mail", handlers.WelcomeMessage{
		FullName: fmt.Sprintf("%s %s", *user.FirstName, *user.LastName),
		Year:     time.Now().UTC().Year(),
	})
	if err != nil {
		log.Printf("Error sending welcome mail: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member created successfully"})
}

func (u *UserHandler) Login(c *gin.Context) {
	var payload LoginPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Validate at least one identifier exists
	if payload.Email == nil && payload.UserName == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email or username is required"})
		return
	}

	// Validate password meets requirements
	if len(payload.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	var existingUser model.User
	var err error

	if !emailRegex.MatchString(*payload.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	err = u.UserRepository.GetByEmail(&existingUser, *payload.Email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid login credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Verify password
	passwordHasher, err := utils.NewPasswordHasher()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	matches, err := passwordHasher.VerifyPassword(*existingUser.Password, payload.Password)
	if err != nil || !matches {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	tokenGenerator, err := jwt.NewSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	accessToken, err := tokenGenerator.GenerateAccessToken(existingUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := tokenGenerator.GenerateRefreshToken(existingUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user": gin.H{
			"email": existingUser.Email,
			"name":  fmt.Sprintf("%s %s", *existingUser.FirstName, *existingUser.LastName),
		},
	})
}
