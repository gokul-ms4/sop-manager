package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gokul-ms4/sop-manager/common/models"
	"github.com/gokul-ms4/sop-manager/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func Register(c echo.Context) error {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	var user models.User
	jsonData := c.FormValue("data")
	err := json.Unmarshal([]byte(jsonData), &user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "unsupported input Format",
		})
	}
	if user.Name == "" || user.Email == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Username, email and password are required",
		})
	}
	user.Email = strings.TrimSpace(user.Email)
	result := config.DB.Unscoped().Where("email=?", user.Email).First(&user)
	if result.RowsAffected > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Email already exists",
		})
	}
	re := regexp.MustCompile(emailRegex)
	if !re.MatchString(user.Email) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid Email",
		})
	}
	if user.PhoneNumber == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Phone Number is required",
		})
	}
	if !phoneRegex.MatchString(user.PhoneNumber) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid Phone Number",
		})
	}
	result = config.DB.Unscoped().Where("phone_number=?", user.PhoneNumber).First(&user)
	if result.RowsAffected > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Phone number already exists",
		})
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to process password",
		})
	}
	user.Password = string(hashed)
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		fileurl, err := config.UploadFile(file)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}
		user.Avatar = fileurl
	}
	if err := config.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Registration failed",
		})
	}
	user.Password = ""
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"response": map[string]any{
			"message": "Registration completed successfully",
			"data":    user,
		},
	})
}

func Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Email and password are required",
		})
	}

	var user models.User
	result := config.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	accessToken, err := generateToken(user.ID, "access", time.Hour*24)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to generate access token",
		})
	}

	refreshToken, err := generateToken(user.ID, "refresh", time.Hour*24*7)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to generate refresh token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

func generateToken(userID int, tokenType string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    tokenType,
		"exp":     time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Refresh token is required",
		})
	}

	// Validate refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "Invalid or expired refresh token",
		})
	}

	claims := token.Claims.(jwt.MapClaims)

	// Make sure it's actually a refresh token
	if claims["type"] != "refresh" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "Invalid token type",
		})
	}

	userID := uint(claims["user_id"].(float64))

	// Check if user still exists and is active
	var user models.User
	result := config.DB.Where("id = ? AND is_active = ?", userID, true).First(&user)
	if result.Error != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "User not found or inactive",
		})
	}

	// Generate new access token only
	accessToken, err := generateToken(user.ID, "access", time.Hour*24)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to generate access token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"access_token": accessToken,
		},
	})
}

func ForgotPassword(c echo.Context) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Email is required",
		})
	}

	var user models.User
	result := config.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		// Don't reveal if email exists or not
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "If this email exists you will receive a reset link shortly",
		})
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expiry := time.Now().Add(time.Hour * 1) // 1 hour expiry

	config.DB.Model(&user).Updates(map[string]interface{}{
		"password_reset_token":  resetToken,
		"password_reset_expiry": expiry,
	})

	// Build reset link
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	go config.SendEmail(config.ForgotPasswordEmailTemplate(user.Name, user.Email, resetLink))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "If this email exists you will receive a reset link shortly",
	})
}

func ResetPassword(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Reset token is required",
		})
	}

	var req struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	if req.Password == "" || req.ConfirmPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Password and confirm password are required",
		})
	}

	if req.Password != req.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Passwords do not match",
		})
	}

	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Password must be at least 6 characters",
		})
	}

	// Find user by token
	var user models.User
	result := config.DB.Where("password_reset_token = ?", token).First(&user)
	if result.Error != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid or expired reset token",
		})
	}

	// Check expiry
	if user.PasswordResetExpiry == nil || time.Now().After(*user.PasswordResetExpiry) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Reset token has expired",
		})
	}

	// Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to process password",
		})
	}

	// Update password and clear reset token
	config.DB.Model(&user).Updates(map[string]interface{}{
		"password":              string(hashed),
		"password_reset_token":  "",
		"password_reset_expiry": nil,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Password reset successfully",
	})
}
