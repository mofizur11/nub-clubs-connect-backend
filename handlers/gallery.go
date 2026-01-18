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

// UploadEventGallery uploads a photo to event gallery
func UploadEventGallery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
		Caption  string `json:"caption"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var galleryID int
	err = database.DB.QueryRow(
		`INSERT INTO event_gallery (event_id, uploaded_by, image_url, caption)
		 VALUES ($1, $2, $3, $4)
		 RETURNING gallery_id`,
		eventID, userID, req.ImageURL, req.Caption,
	).Scan(&galleryID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upload gallery image")
		return
	}

	response := gin.H{
		"gallery_id": galleryID,
		"image_url":  req.ImageURL,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Gallery image uploaded successfully", response)
}

// GetEventGallery retrieves all photos from an event
func GetEventGallery(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			eg.gallery_id, eg.image_url, eg.caption, eg.uploaded_at, eg.uploaded_by,
			u.first_name || ' ' || u.last_name as uploaded_by_name
		 FROM event_gallery eg
		 LEFT JOIN users u ON eg.uploaded_by = u.user_id
		 WHERE eg.event_id = $1
		 ORDER BY eg.uploaded_at DESC`,
		eventID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch gallery")
		return
	}
	defer rows.Close()

	var gallery []models.EventGallery

	for rows.Next() {
		var item models.EventGallery
		var uploadedByName sql.NullString
		err := rows.Scan(&item.GalleryID, &item.ImageURL, &item.Caption, &item.UploadedAt, &item.UploadedBy, &uploadedByName)
		if err != nil {
			continue
		}
		if uploadedByName.Valid {
			item.UploadedByName = uploadedByName.String
		}
		gallery = append(gallery, item)
	}

	utils.SuccessResponse(c, http.StatusOK, "Event gallery retrieved", gallery)
}

// DeleteGalleryImage deletes a photo from event gallery
func DeleteGalleryImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	galleryID, err := strconv.Atoi(c.Param("galleryId"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid gallery ID")
		return
	}

	// Check if user is the uploader or admin
	var uploadedBy int
	err = database.DB.QueryRow(
		`SELECT uploaded_by FROM event_gallery WHERE gallery_id = $1`,
		galleryID,
	).Scan(&uploadedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Gallery image not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to verify ownership")
		return
	}

	role, _ := c.Get("role")
	if uploadedBy != userID.(int) && role != "system_admin" {
		utils.ForbiddenResponse(c, "You can only delete your own uploads")
		return
	}

	_, err = database.DB.Exec(
		`DELETE FROM event_gallery WHERE gallery_id = $1`,
		galleryID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete gallery image")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Gallery image deleted successfully", nil)
}
