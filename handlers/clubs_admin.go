package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// UpdateClub updates club details (admin only)
func UpdateClub(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can update clubs")
		return
	}

	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	type updateClubRequest struct {
		ClubName      *string `json:"club_name"`
		ClubCode      *string `json:"club_code"`
		Description   *string `json:"description"`
		LogoURL       *string `json:"logo_url"`
		CoverImageURL *string `json:"cover_image_url"`
		FoundedDate   *string `json:"founded_date"`
		Email         *string `json:"email"`
	}

	var req updateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	sets := []string{}
	args := []interface{}{}
	idx := 1

	if req.ClubName != nil {
		sets = append(sets, "club_name = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.ClubName))
		idx++
	}
	if req.ClubCode != nil {
		sets = append(sets, "club_code = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.ClubCode))
		idx++
	}
	if req.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.Description))
		idx++
	}
	if req.LogoURL != nil {
		sets = append(sets, "logo_url = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.LogoURL))
		idx++
	}
	if req.CoverImageURL != nil {
		sets = append(sets, "cover_image_url = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.CoverImageURL))
		idx++
	}
	if req.FoundedDate != nil {
		var tPtr *time.Time
		if *req.FoundedDate != "" {
			if t, perr := time.Parse(time.RFC3339, *req.FoundedDate); perr == nil {
				tPtr = &t
			}
		}
		sets = append(sets, "founded_date = $"+strconv.Itoa(idx))
		args = append(args, tPtr)
		idx++
	}
	if req.Email != nil {
		sets = append(sets, "email = $"+strconv.Itoa(idx))
		args = append(args, strings.TrimSpace(*req.Email))
		idx++
	}

	if len(sets) == 0 {
		utils.BadRequestResponse(c, "No fields to update")
		return
	}

	sets = append(sets, "updated_at = $"+strconv.Itoa(idx))
	args = append(args, time.Now())
	idx++

	query := "UPDATE clubs SET " + strings.Join(sets, ", ") + " WHERE club_id = $" + strconv.Itoa(idx)
	args = append(args, clubID)

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update club")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Club updated successfully", nil)
}

// RemoveModerator removes a moderator from a club (admin only)
func RemoveModerator(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can remove moderators")
		return
	}

	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	_, err = database.DB.Exec(
		`DELETE FROM club_moderators WHERE user_id = $1 AND club_id = $2`,
		userID, clubID,
	)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to remove moderator")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Moderator removed successfully", nil)
}
