package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// GetUserActivityLog retrieves activity log for a specific user
func GetUserActivityLog(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT log_id, action, entity_type, entity_id, details, ip_address, created_at
		 FROM activity_log
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 100`,
		userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch activity log")
		return
	}
	defer rows.Close()

	var logs []models.ActivityLog

	for rows.Next() {
		var log models.ActivityLog
		err := rows.Scan(&log.LogID, &log.Action, &log.EntityType, &log.EntityID, &log.Details, &log.IPAddress, &log.CreatedAt)
		if err != nil {
			continue
		}
		log.UserID = userID
		logs = append(logs, log)
	}

	utils.SuccessResponse(c, http.StatusOK, "Activity log retrieved", logs)
}

// GetCurrentUserActivityLog retrieves activity log for the current authenticated user
func GetCurrentUserActivityLog(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	rows, err := database.DB.Query(
		`SELECT log_id, action, entity_type, entity_id, details, ip_address, created_at
		 FROM activity_log
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 100`,
		userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch activity log")
		return
	}
	defer rows.Close()

	var logs []models.ActivityLog

	for rows.Next() {
		var log models.ActivityLog
		err := rows.Scan(&log.LogID, &log.Action, &log.EntityType, &log.EntityID, &log.Details, &log.IPAddress, &log.CreatedAt)
		if err != nil {
			continue
		}
		log.UserID = userID.(int)
		logs = append(logs, log)
	}

	utils.SuccessResponse(c, http.StatusOK, "Activity log retrieved", logs)
}

// GetAllActivityLogs retrieves all activity logs (admin only)
func GetAllActivityLogs(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view all activity logs")
		return
	}

	rows, err := database.DB.Query(
		`SELECT log_id, user_id, action, entity_type, entity_id, details, ip_address, created_at
		 FROM activity_log
		 ORDER BY created_at DESC
		 LIMIT 500`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch activity logs")
		return
	}
	defer rows.Close()

	var logs []models.ActivityLog

	for rows.Next() {
		var log models.ActivityLog
		err := rows.Scan(&log.LogID, &log.UserID, &log.Action, &log.EntityType, &log.EntityID, &log.Details, &log.IPAddress, &log.CreatedAt)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	utils.SuccessResponse(c, http.StatusOK, "Activity logs retrieved", logs)
}
