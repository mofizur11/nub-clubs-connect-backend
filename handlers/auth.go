package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// Register handles user registration
func Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to process password")
		return
	}

	// Insert user into database
	var userID int
	var email string
	var firstName string
	var lastName string
	var studentID string
	var createdAt time.Time
	var updatedAt time.Time

	err = database.DB.QueryRow(
		`INSERT INTO users (student_id, email, password_hash, first_name, last_name, role)
		 VALUES ($1, $2, $3, $4, $5, 'student')
		 RETURNING user_id, student_id, email, first_name, last_name, created_at, updated_at`,
		req.StudentID, req.Email, hashedPassword, req.FirstName, req.LastName,
	).Scan(&userID, &studentID, &email, &firstName, &lastName, &createdAt, &updatedAt)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			utils.ConflictResponse(c, "Email already registered")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to register user")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(userID, email, "student")
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate token")
		return
	}

	response := gin.H{
		"user_id": userID,
		"student_id": studentID,
		"email":   email,
		"first_name": firstName,
		"last_name": lastName,
		"role": "student",
		"created_at": createdAt,
		"updated_at": updatedAt,
		"token": token,
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", response)
}

// Login handles user login
func Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var user models.User

	err := database.DB.QueryRow(
		`SELECT user_id, student_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at
		 FROM users
		 WHERE email = $1`,
		req.Email,
	).Scan(&user.UserID, &user.StudentID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.UnauthorizedResponse(c, "Invalid email or password")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to authenticate user")
		return
	}

	if !user.IsActive {
		utils.UnauthorizedResponse(c, "User account is inactive")
		return
	}

	// Verify password
	if !utils.VerifyPassword(user.PasswordHash, req.Password) {
		utils.UnauthorizedResponse(c, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate token")
		return
	}

	response := gin.H{
		"user_id": user.UserID,
		"student_id": user.StudentID,
		"email": user.Email,
		"first_name": user.FirstName,
		"last_name": user.LastName,
		"role": user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		"token": token,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// GetProfile gets the current user's profile
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var user models.User

	err := database.DB.QueryRow(
		`SELECT user_id, student_id, email, first_name, last_name, role, phone, profile_picture_url, is_active, created_at, updated_at
		 FROM users
		 WHERE user_id = $1`,
		userID,
	).Scan(&user.UserID, &user.StudentID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.Phone, &user.ProfilePictureURL, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to fetch user profile")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User profile retrieved", user)
}

// UpdateProfile updates the current user's profile
func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	_, err := database.DB.Exec(
		`UPDATE users
		 SET first_name = COALESCE(NULLIF($1, ''), first_name),
		     last_name = COALESCE(NULLIF($2, ''), last_name),
		     phone = COALESCE(NULLIF($3, ''), phone),
		     profile_picture_url = COALESCE(NULLIF($4, ''), profile_picture_url),
		     updated_at = CURRENT_TIMESTAMP
		 WHERE user_id = $5`,
		req.FirstName, req.LastName, req.Phone, req.ProfilePictureURL, userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", nil)
}

// ForgotPassword handles password reset request
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var userID int
	err := database.DB.QueryRow(
		`SELECT user_id FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists
			utils.SuccessResponse(c, http.StatusOK, "If email exists, password reset link has been sent", nil)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to process request")
		return
	}

	// Generate reset token
	token, err := utils.GenerateResetToken()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate reset token")
		return
	}

	// Store token in database
	_, err = database.DB.Exec(
		`INSERT INTO password_reset_tokens (user_id, token, expires_at)
		 VALUES ($1, $2, CURRENT_TIMESTAMP + INTERVAL '1 hour')`,
		userID, token,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create reset token")
		return
	}

	// In production, send email with reset link
	utils.SuccessResponse(c, http.StatusOK, "If email exists, password reset link has been sent", nil)
}

// ResetPassword handles password reset
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	// Validate token
	var userID int
	err := database.DB.QueryRow(
		`SELECT user_id FROM password_reset_tokens
		 WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP AND is_used = FALSE`,
		req.Token,
	).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.UnauthorizedResponse(c, "Invalid or expired reset token")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to validate token")
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to process password")
		return
	}

	// Update password
	_, err = database.DB.Exec(
		`UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2`,
		hashedPassword, userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update password")
		return
	}

	// Mark token as used
	_, err = database.DB.Exec(
		`UPDATE password_reset_tokens SET is_used = TRUE WHERE token = $1`,
		req.Token,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to mark token as used")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password reset successfully", nil)
}
