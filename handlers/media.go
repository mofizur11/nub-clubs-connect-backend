package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// UploadNewsMedia uploads media to a news post
func UploadNewsMedia(c *gin.Context) {
	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid news ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req struct {
		MediaType    string `json:"media_type" binding:"required"`
		MediaURL     string `json:"media_url" binding:"required"`
		Caption      string `json:"caption"`
		DisplayOrder int    `json:"display_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var mediaID int
	err = database.DB.QueryRow(
		`INSERT INTO news_media (news_id, media_type, media_url, caption, display_order, uploaded_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING media_id`,
		newsID, req.MediaType, req.MediaURL, req.Caption, req.DisplayOrder, userID,
	).Scan(&mediaID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upload media")
		return
	}

	response := gin.H{
		"media_id": mediaID,
		"media_url": req.MediaURL,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Media uploaded successfully", response)
}

// GetNewsMedia retrieves all media for a news post
func GetNewsMedia(c *gin.Context) {
	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid news ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT media_id, media_type, media_url, caption, display_order, uploaded_at
		 FROM news_media
		 WHERE news_id = $1
		 ORDER BY display_order ASC`,
		newsID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch media")
		return
	}
	defer rows.Close()

	var media []models.NewsMedia

	for rows.Next() {
		var item models.NewsMedia
		err := rows.Scan(&item.MediaID, &item.MediaType, &item.MediaURL, &item.Caption, &item.DisplayOrder, &item.UploadedAt)
		if err != nil {
			continue
		}
		media = append(media, item)
	}

	utils.SuccessResponse(c, http.StatusOK, "News media retrieved", media)
}

// DeleteNewsMedia deletes media from a news post
func DeleteNewsMedia(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	mediaID, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid media ID")
		return
	}

	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can delete media")
		return
	}

	_, err = database.DB.Exec(
		`DELETE FROM news_media WHERE media_id = $1`,
		mediaID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete media")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Media deleted successfully", nil)
}
