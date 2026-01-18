package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// AdminListUsers returns a paginated list of users with optional filters
func AdminListUsers(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can manage users")
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	roleFilter := strings.TrimSpace(c.Query("role"))
	isActiveStr := strings.TrimSpace(c.Query("is_active"))
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if q != "" {
		where = append(where, "(LOWER(first_name) LIKE $"+strconv.Itoa(argIdx)+" OR LOWER(last_name) LIKE $"+strconv.Itoa(argIdx)+" OR LOWER(email) LIKE $"+strconv.Itoa(argIdx)+" OR LOWER(student_id) LIKE $"+strconv.Itoa(argIdx)+")")
		args = append(args, "%"+strings.ToLower(q)+"%")
		argIdx++
	}
	if roleFilter != "" {
		where = append(where, "role = $"+strconv.Itoa(argIdx))
		args = append(args, roleFilter)
		argIdx++
	}
	if isActiveStr != "" {
		isActive := strings.EqualFold(isActiveStr, "true") || isActiveStr == "1"
		where = append(where, "is_active = $"+strconv.Itoa(argIdx))
		args = append(args, isActive)
		argIdx++
	}

	query := `SELECT user_id, student_id, email, first_name, last_name, role, phone, profile_picture_url, is_active, created_at, updated_at
	          FROM users`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIdx) + " OFFSET $" + strconv.Itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to list users")
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		var userID int
		var studentID, email, firstName, lastName, dbRole, phone, profileURL sql.NullString
		var isActive sql.NullBool
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&userID, &studentID, &email, &firstName, &lastName, &dbRole, &phone, &profileURL, &isActive, &createdAt, &updatedAt); err != nil {
			continue
		}
		u.UserID = userID
		u.StudentID = models.NullString(studentID)
		u.Email = models.NullString(email)
		u.FirstName = models.NullString(firstName)
		u.LastName = models.NullString(lastName)
		u.Role = models.NullString(dbRole)
		u.Phone = models.NullString(phone)
		u.ProfilePictureURL = models.NullString(profileURL)
		if isActive.Valid {
			u.IsActive = isActive.Bool
		} else {
			u.IsActive = false
		}
		u.CreatedAt = createdAt
		u.UpdatedAt = updatedAt
		users = append(users, u)
	}

	utils.SuccessResponse(c, http.StatusOK, "Users retrieved", users)
}

// AdminGetUserByID returns details for a single user
func AdminGetUserByID(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can manage users")
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}
	var u models.User
	var userID int
	var studentID, email, firstName, lastName, dbRole, phone, profileURL sql.NullString
	var isActive sql.NullBool
	var createdAt, updatedAt time.Time
	err = database.DB.QueryRow(`SELECT user_id, student_id, email, first_name, last_name, role, phone, profile_picture_url, is_active, created_at, updated_at
		FROM users WHERE user_id = $1`, id).Scan(&userID, &studentID, &email, &firstName, &lastName, &dbRole, &phone, &profileURL, &isActive, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to fetch user")
		return
	}
	u.UserID = userID
	u.StudentID = models.NullString(studentID)
	u.Email = models.NullString(email)
	u.FirstName = models.NullString(firstName)
	u.LastName = models.NullString(lastName)
	u.Role = models.NullString(dbRole)
	u.Phone = models.NullString(phone)
	u.ProfilePictureURL = models.NullString(profileURL)
	if isActive.Valid {
		u.IsActive = isActive.Bool
	} else {
		u.IsActive = false
	}
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	utils.SuccessResponse(c, http.StatusOK, "User retrieved", u)
}

// adminUpdateUserRequest are optional fields that admin can change
type adminUpdateUserRequest struct {
	StudentID         *string `json:"student_id"`
	Email             *string `json:"email"`
	FirstName         *string `json:"first_name"`
	LastName          *string `json:"last_name"`
	Phone             *string `json:"phone"`
	ProfilePictureURL *string `json:"profile_picture_url"`
	IsActive          *bool   `json:"is_active"`
}

// AdminUpdateUser updates user fields
func AdminUpdateUser(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can manage users")
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	var req adminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	sets := []string{}
	args := []interface{}{}
	idx := 1

	if req.StudentID != nil {
		sets = append(sets, "student_id = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.StudentID))
		idx++
	}
	if req.Email != nil {
		sets = append(sets, "email = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.Email))
		idx++
	}
	if req.FirstName != nil {
		sets = append(sets, "first_name = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.FirstName))
		idx++
	}
	if req.LastName != nil {
		sets = append(sets, "last_name = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.LastName))
		idx++
	}
	if req.Phone != nil {
		sets = append(sets, "phone = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.Phone))
		idx++
	}
	if req.ProfilePictureURL != nil {
		sets = append(sets, "profile_picture_url = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.ProfilePictureURL))
		idx++
	}
	if req.IsActive != nil {
		sets = append(sets, "is_active = $"+strconv.Itoa(idx))
		args = append(args, *req.IsActive)
		idx++
	}

	if len(sets) == 0 {
		utils.BadRequestResponse(c, "No fields to update")
		return
	}

	sets = append(sets, "updated_at = $"+strconv.Itoa(idx))
	args = append(args, time.Now())
	idx++

	query := "UPDATE users SET " + strings.Join(sets, ", ") + " WHERE user_id = $" + strconv.Itoa(idx)
	args = append(args, id)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", nil)
}

// AdminDeleteUser deletes a user
func AdminDeleteUser(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can manage users")
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	_, err = database.DB.Exec("DELETE FROM users WHERE user_id = $1", id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}

// changeRoleRequest is the payload for role update
type changeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// AdminChangeUserRole changes a user's role
func AdminChangeUserRole(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can manage users")
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}
	var req changeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}
	newRole := strings.TrimSpace(req.Role)
	if newRole != "student" && newRole != "club_moderator" && newRole != "system_admin" {
		utils.BadRequestResponse(c, "Invalid role")
		return
	}

	_, err = database.DB.Exec("UPDATE users SET role = $1, updated_at = $2 WHERE user_id = $3", newRole, time.Now(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to change user role")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User role updated successfully", nil)
}
