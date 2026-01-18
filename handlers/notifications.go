package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// GetUserNotifications retrieves all notifications for the current user
func GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	rows, err := database.DB.Query(
		`SELECT notification_id, title, message, notification_type, related_entity_type, related_entity_id, is_read, created_at
		 FROM notifications
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 50`,
		userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch notifications")
		return
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var notif models.Notification
		err := rows.Scan(&notif.NotificationID, &notif.Title, &notif.Message, &notif.NotificationType,
			&notif.RelatedEntityType, &notif.RelatedEntityID, &notif.IsRead, &notif.CreatedAt)
		if err != nil {
			continue
		}
		notifications = append(notifications, notif)
	}

	utils.SuccessResponse(c, http.StatusOK, "Notifications retrieved", notifications)
}

// GetUnreadNotificationCount retrieves count of unread notifications
func GetUnreadNotificationCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var count int
	err := database.DB.QueryRow(
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	).Scan(&count)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch notification count")
		return
	}

	response := gin.H{
		"unread_count": count,
	}

	utils.SuccessResponse(c, http.StatusOK, "Unread notification count retrieved", response)
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	notificationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid notification ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE notifications SET is_read = TRUE WHERE notification_id = $1 AND user_id = $2`,
		notificationID, userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to mark notification as read")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllNotificationsAsRead marks all notifications as read for current user
func MarkAllNotificationsAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	_, err := database.DB.Exec(
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to mark notifications as read")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All notifications marked as read", nil)
}

// DeleteNotification deletes a notification
func DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	notificationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid notification ID")
		return
	}

	_, err = database.DB.Exec(
		`DELETE FROM notifications WHERE notification_id = $1 AND user_id = $2`,
		notificationID, userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete notification")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification deleted", nil)
}
