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
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserController struct {
	UserRepository           repositories.UserRepository
	ForgotPasswordRepository repositories.ForgotPasswordRepository
}

func NewUserHandler(
	userRepository *repositories.UserRepository,
	forgotPasswordRepository *repositories.ForgotPasswordRepository,
) *UserController {
	return &UserController{
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

func (u *UserController) Create(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (u *UserController) CompleteAccount(c *gin.Context) {
	userIDStr := c.Param("userId")
	documentID := c.Query("documentId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var userInput struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Password  string `json:"password"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		log.Printf("Error binding JSON: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ok, err := nameRegex.MatchString(userInput.FirstName)
	if err != nil || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid first name"})
		return
	}

	ok, err = nameRegex.MatchString(userInput.LastName)
	if err != nil || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last name"})
		return
	}

	if !passwordRegex.MatchString(userInput.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password format"})
		return
	}

	var user model.User
	if err := u.UserRepository.GetById(&user, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Error fetching user: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	hasher, err := utils.NewPasswordHasher()
	if err != nil {
		log.Printf("Error creating password hasher: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	hashedPassword, err := hasher.HashPassword(userInput.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	user.FirstName = &userInput.FirstName
	user.LastName = &userInput.LastName
	user.Password = &hashedPassword

	if err := u.UserRepository.Update(&user, userID); err != nil {
		log.Printf("Error updating user: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete account setup"})
		return
	}

	err = handlers.SendMail(user.Email, "welcome", "Welcome to FileEditor", handlers.WelcomeMessage{
		FullName: fmt.Sprintf("%s %s", userInput.FirstName, userInput.LastName),
		Year:     time.Now().UTC().Year(),
	})
	if err != nil {
		log.Printf("Error sending welcome email: %s", err)
	}

	// redirectURL := fmt.Sprintf("%s/get-document/%s", envVars.FE_ROOT_URL, documentID)
	// c.Redirect(http.StatusFound, redirectURL)

	tokenGenerator, err := jwt.NewSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	accessToken, err := tokenGenerator.GenerateAccessToken(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := tokenGenerator.GenerateRefreshToken(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Account setup complete",
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"redirectTo":   fmt.Sprintf("/get-document/%s", documentID),
	})
}

func (u *UserController) Login(c *gin.Context) {
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

func (u *UserController) UploadProfilePicture(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type"})
		return
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form data"})
		return
	}

	// Get profile picture file
	files, exists := form.File["profilePicture"]
	if !exists || len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile picture file is required"})
		return
	}

	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer file.Close()

	// Delete existing photo if exists
	if userDetails.ProfilePhoto != nil {
		_, err := repositories.CloudinaryDelete(userDetails.ProfilePhoto.Public_ID, repositories.ImageResource)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete existing profile photo"})
			return
		}
	}

	// Upload new photo
	uploaded, err := repositories.CloudinaryUploaderStream(file, fileHeader.Filename, repositories.ImageResource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload profile photo"})
		return
	}

	// Update user record
	userDetails.ProfilePhoto = &model.Media{
		Public_ID:  uploaded.PublicID,
		Secure_URL: uploaded.SecureURL,
	}

	if err := u.UserRepository.Update(&userDetails, userDetails.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "profile photo uploaded successfully",
		"imageUrl": uploaded.SecureURL,
	})
}

func (u *UserController) ForgotPassword(c *gin.Context) {
	var user model.User

	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Error binding JSON: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !emailRegex.MatchString(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var existingUser model.User
	err := u.UserRepository.GetByEmail(&existingUser, user.Email)

	if err != nil {
		log.Printf("Error:%s", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found"})
			return
		}
		log.Printf("Error getting user: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resetCode := utils.NewCodeGenerator().GenerateEmailVerificationCode(6)
	forgotPassword := &model.ForgotPassword{
		Email:     existingUser.Email,
		ResetCode: resetCode,
	}

	if err := u.ForgotPasswordRepository.Create(forgotPassword); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	err = handlers.SendMail(user.Email, "forgotPassword", "Password Reset Mail", handlers.PasswordResetCode{
		FullName:  fmt.Sprintf("%s %s", *existingUser.FirstName, *existingUser.LastName),
		ResetCode: resetCode,
		Year:      time.Now().UTC().Year(),
	})
	if err != nil {
		log.Printf("Error:%s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending forgot password mail"})
		return
	}
}

func (u *UserController) VerifyResetCode(c *gin.Context) {
	var payload ResetCodePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !codeRegex.MatchString(payload.ResetCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code"})
	}

	var forgotPassword model.ForgotPassword
	if err := u.ForgotPasswordRepository.GetOneByCode(payload.ResetCode, &forgotPassword); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := u.UserRepository.GetByEmail(&user, forgotPassword.Email); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email does not exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := u.ForgotPasswordRepository.Delete(&forgotPassword, forgotPassword.ID); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	tokenGenerator, err := jwt.NewSession()
	if err != nil {
		log.Printf("Error:%s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := tokenGenerator.GenerateAccessToken(user.Email)
	if err != nil {
		log.Printf("Error:%s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (u *UserController) ResetPassword(c *gin.Context) {
	// Extract the authenticated user from the context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
		return
	}

	var payload struct {
		NewPassword string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	passwordHasher, err := utils.NewPasswordHasher()
	if err != nil {
		log.Printf("Error creating password hasher: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	hashedPassword, err := passwordHasher.HashPassword(payload.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	userDetails.Password = &hashedPassword
	if err := u.UserRepository.Update(&userDetails, userDetails.ID); err != nil {
		log.Printf("Error updating password: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (u *UserController) GenerateAccessToken(c *gin.Context) {
	var payload struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	tokenGenerator, err := jwt.NewSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	accessToken, err := tokenGenerator.VerifyRefreshToken(payload.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": accessToken})
}

func (u *UserController) Profile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		log.Printf("Error: user not found in context")
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid session"})
		return
	}

	userDetails, ok := user.(model.User)
	if !ok {
		log.Printf("Error: invalid user type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
		return
	}

	// Fetch the latest user data from the database
	var freshMember model.User
	err := u.UserRepository.GetById(&freshMember, userDetails.ID)
	if err != nil {
		log.Printf("Error fetching latest user data: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch profile"})
		return
	}

	newDetails := model.User{
		FirstName: freshMember.FirstName,
		LastName:  freshMember.LastName,
		Email:     freshMember.Email,

		ProfilePhoto: freshMember.ProfilePhoto,
	}

	c.JSON(http.StatusOK, gin.H{"user": newDetails})
}
