package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// CreateSystemAnnouncement creates a system-wide announcement
func CreateSystemAnnouncement(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can create announcements")
		return
	}

	var req struct {
		Title     string `json:"title" binding:"required"`
		Content   string `json:"content" binding:"required"`
		Priority  string `json:"priority"`
		ExpiresAt *string `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var announcementID int
	err := database.DB.QueryRow(
		`INSERT INTO system_announcements (created_by, title, content, priority, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING announcement_id`,
		userID, req.Title, req.Content, req.Priority, req.ExpiresAt,
	).Scan(&announcementID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create announcement")
		return
	}

	response := gin.H{
		"announcement_id": announcementID,
		"title":           req.Title,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Announcement created successfully", response)
}

// GetSystemAnnouncements retrieves all active system announcements
func GetSystemAnnouncements(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT 
			sa.announcement_id, sa.title, sa.content, sa.priority, sa.created_at, sa.expires_at,
			u.first_name || ' ' || u.last_name as created_by_name
		 FROM system_announcements sa
		 LEFT JOIN users u ON sa.created_by = u.user_id
		 WHERE sa.is_active = TRUE AND (sa.expires_at IS NULL OR sa.expires_at > CURRENT_TIMESTAMP)
		 ORDER BY sa.created_at DESC`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch announcements")
		return
	}
	defer rows.Close()

	var announcements []models.SystemAnnouncement

	for rows.Next() {
		var ann models.SystemAnnouncement
		var createdByName sql.NullString
		err := rows.Scan(&ann.AnnouncementID, &ann.Title, &ann.Content, &ann.Priority, &ann.CreatedAt, &ann.ExpiresAt, &createdByName)
		if err != nil {
			continue
		}
		if createdByName.Valid {
			ann.CreatedByName = createdByName.String
		}
		announcements = append(announcements, ann)
	}

	utils.SuccessResponse(c, http.StatusOK, "System announcements retrieved", announcements)
}

// GetAnnouncementDetails retrieves a specific announcement
func GetAnnouncementDetails(c *gin.Context) {
	announcementID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid announcement ID")
		return
	}

	var ann models.SystemAnnouncement
	var createdByName sql.NullString

	err = database.DB.QueryRow(
		`SELECT 
			sa.announcement_id, sa.title, sa.content, sa.priority, sa.is_active, sa.created_at, sa.expires_at,
			u.first_name || ' ' || u.last_name as created_by_name
		 FROM system_announcements sa
		 LEFT JOIN users u ON sa.created_by = u.user_id
		 WHERE sa.announcement_id = $1`,
		announcementID,
	).Scan(&ann.AnnouncementID, &ann.Title, &ann.Content, &ann.Priority, &ann.IsActive, &ann.CreatedAt, &ann.ExpiresAt, &createdByName)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Announcement not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to fetch announcement")
		return
	}

	if createdByName.Valid {
		ann.CreatedByName = createdByName.String
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement retrieved", ann)
}

// UpdateSystemAnnouncement updates a system announcement
func UpdateSystemAnnouncement(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can update announcements")
		return
	}

	announcementID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid announcement ID")
		return
	}

	var req struct {
		Title     string `json:"title"`
		Content   string `json:"content"`
		Priority  string `json:"priority"`
		IsActive  *bool  `json:"is_active"`
		ExpiresAt *string `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE system_announcements
		 SET title = COALESCE(NULLIF($1, ''), title),
		     content = COALESCE(NULLIF($2, ''), content),
		     priority = COALESCE(NULLIF($3, ''), priority),
		     is_active = COALESCE($4, is_active),
		     expires_at = COALESCE($5::TIMESTAMP, expires_at)
		 WHERE announcement_id = $6`,
		req.Title, req.Content, req.Priority, req.IsActive, req.ExpiresAt, announcementID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update announcement")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement updated successfully", nil)
}

// DeleteSystemAnnouncement deletes a system announcement
func DeleteSystemAnnouncement(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can delete announcements")
		return
	}

	announcementID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid announcement ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE system_announcements SET is_active = FALSE WHERE announcement_id = $1`,
		announcementID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete announcement")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Announcement deleted successfully", nil)
}
