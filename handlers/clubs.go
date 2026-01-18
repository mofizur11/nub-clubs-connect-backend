package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// CreateClub creates a new club
func CreateClub(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can create clubs")
		return
	}

	var req models.CreateClubRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var clubID int
	err := database.DB.QueryRow(
		`INSERT INTO clubs (club_name, club_code, description, logo_url, cover_image_url, founded_date, email)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING club_id`,
		req.ClubName, req.ClubCode, req.Description, req.LogoURL, req.CoverImageURL, req.FoundedDate, req.Email,
	).Scan(&clubID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create club")
		return
	}

	// Log activity
	LogActivity(userID.(int), "club_created", "club", clubID, nil)

	response := gin.H{
		"club_id": clubID,
		"club_name": req.ClubName,
		"club_code": req.ClubCode,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Club created successfully", response)
}

// GetAllClubs retrieves all clubs
func GetAllClubs(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT club_id, club_name, club_code, description, logo_url, cover_image_url, is_active, created_at, updated_at
		 FROM clubs
		 WHERE is_active = TRUE
		 ORDER BY club_name`,
	)

	if err != nil {
		fmt.Printf("DEBUG: Query error: %v\n", err)
		utils.InternalServerErrorResponse(c, "Failed to fetch clubs")
		return
	}
	defer rows.Close()

	clubs := make([]models.Club, 0)

	for rows.Next() {
		var club models.Club
		var logoURL sql.NullString
		var coverImageURL sql.NullString
		err := rows.Scan(&club.ClubID, &club.ClubName, &club.ClubCode, &club.Description, &logoURL, &coverImageURL, &club.IsActive, &club.CreatedAt, &club.UpdatedAt)
		if err != nil {
			fmt.Printf("DEBUG: Scan error: %v\n", err)
			continue
		}
		if logoURL.Valid {
			club.LogoURL = logoURL.String
		}
		if coverImageURL.Valid {
			club.CoverImageURL = coverImageURL.String
		}
		clubs = append(clubs, club)
	}

	fmt.Printf("DEBUG: Clubs found: %d\n", len(clubs))
	utils.SuccessResponse(c, http.StatusOK, "Clubs retrieved successfully", clubs)
}

// GetClubDetails retrieves club details with member and event counts
func GetClubDetails(c *gin.Context) {
	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	var club models.Club
	var memberCount int
	var upcomingEvents int
	var descNS sql.NullString
	var logoNS sql.NullString
	var coverNS sql.NullString
	var foundedNT sql.NullTime
	var emailNS sql.NullString

	err = database.DB.QueryRow(
		`SELECT 
			c.club_id, c.club_name, c.club_code, c.description, c.logo_url, c.cover_image_url, c.founded_date, c.email, c.is_active, c.created_at, c.updated_at,
			COUNT(DISTINCT cm.user_id) as member_count,
			COUNT(DISTINCT e.event_id) FILTER (WHERE e.status = 'approved') as upcoming_events
		 FROM clubs c
		 LEFT JOIN club_members cm ON c.club_id = cm.club_id AND cm.is_active = TRUE
		 LEFT JOIN events e ON c.club_id = e.club_id
		 WHERE c.club_id = $1
		 GROUP BY c.club_id`,
		clubID,
	).Scan(&club.ClubID, &club.ClubName, &club.ClubCode, &descNS, &logoNS, &coverNS, &foundedNT, &emailNS, &club.IsActive, &club.CreatedAt, &club.UpdatedAt, &memberCount, &upcomingEvents)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Club not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to fetch club details")
		return
	}

	club.Description = models.NullString(descNS)
	club.LogoURL = models.NullString(logoNS)
	club.CoverImageURL = models.NullString(coverNS)
	club.FoundedDate = models.NullTime(foundedNT)
	club.Email = models.NullString(emailNS)
	club.MemberCount = memberCount
	club.UpcomingEvents = upcomingEvents

	utils.SuccessResponse(c, http.StatusOK, "Club details retrieved", club)
}

// GetClubMembers retrieves all members of a club
func GetClubMembers(c *gin.Context) {
	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			u.user_id, u.first_name, u.last_name, u.email, u.profile_picture_url,
			cm.role, cm.joined_date
		 FROM club_members cm
		 JOIN users u ON cm.user_id = u.user_id
		 WHERE cm.club_id = $1 AND cm.is_active = TRUE
		 ORDER BY cm.joined_date DESC`,
		clubID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch club members")
		return
	}
	defer rows.Close()

	type Member struct {
		UserID             int    `json:"user_id"`
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		Email              string `json:"email"`
		ProfilePictureURL  string `json:"profile_picture_url"`
		Role               string `json:"role"`
		JoinedDate         string `json:"joined_date"`
	}

	members := make([]Member, 0)

	for rows.Next() {
		var member Member
		var lastNameNS sql.NullString
		var emailNS sql.NullString
		var profileNS sql.NullString
		var joinedNT sql.NullTime
		err := rows.Scan(&member.UserID, &member.FirstName, &lastNameNS, &emailNS, &profileNS, &member.Role, &joinedNT)
		if err != nil {
			continue
		}
		member.LastName = models.NullString(lastNameNS)
		member.Email = models.NullString(emailNS)
		member.ProfilePictureURL = models.NullString(profileNS)
		if joinedNT.Valid {
			member.JoinedDate = joinedNT.Time.Format(time.RFC3339)
		} else {
			member.JoinedDate = ""
		}
		members = append(members, member)
	}

	utils.SuccessResponse(c, http.StatusOK, "Club members retrieved", members)
}

// JoinClub adds a user to a club
func JoinClub(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	_, err = database.DB.Exec(
		`INSERT INTO club_members (user_id, club_id, role)
		 VALUES ($1, $2, 'member')
		 ON CONFLICT (user_id, club_id) DO UPDATE SET is_active = TRUE, joined_date = CURRENT_TIMESTAMP`,
		userID, clubID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to join club")
		return
	}

	// Log activity
	LogActivity(userID.(int), "club_joined", "club", clubID, nil)

	utils.SuccessResponse(c, http.StatusOK, "Successfully joined club", nil)
}

// LeaveClub removes a user from a club
func LeaveClub(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE club_members SET is_active = FALSE WHERE user_id = $1 AND club_id = $2`,
		userID, clubID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to leave club")
		return
	}

	// Log activity
	LogActivity(userID.(int), "club_left", "club", clubID, nil)

	utils.SuccessResponse(c, http.StatusOK, "Successfully left club", nil)
}

// GetUserClubs retrieves all clubs a user is a member of
func GetUserClubs(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			c.club_id, c.club_name, c.club_code, c.description, c.logo_url,
			cm.role, cm.joined_date
		 FROM club_members cm
		 JOIN clubs c ON cm.club_id = c.club_id
		 WHERE cm.user_id = $1 AND cm.is_active = TRUE
		 ORDER BY cm.joined_date DESC`,
		userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch user clubs")
		return
	}
	defer rows.Close()

	type ClubInfo struct {
		ClubID       int    `json:"club_id"`
		ClubName     string `json:"club_name"`
		ClubCode     string `json:"club_code"`
		Description  string `json:"description"`
		LogoURL      string `json:"logo_url"`
		Role         string `json:"role"`
		JoinedDate   string `json:"joined_date"`
	}

	clubs := make([]ClubInfo, 0)

	for rows.Next() {
		var club ClubInfo
		err := rows.Scan(&club.ClubID, &club.ClubName, &club.ClubCode, &club.Description, &club.LogoURL, &club.Role, &club.JoinedDate)
		if err != nil {
			continue
		}
		clubs = append(clubs, club)
	}

	utils.SuccessResponse(c, http.StatusOK, "User clubs retrieved", clubs)
}

// AssignModerator assigns a moderator to a club
func AssignModerator(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can assign moderators")
		return
	}

	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	var req struct {
		UserID int `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	_, err = database.DB.Exec(
		`INSERT INTO club_moderators (user_id, club_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, club_id) DO NOTHING`,
		req.UserID, clubID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to assign moderator")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Moderator assigned successfully", nil)
}

// GetClubModerators retrieves all moderators for a club
func GetClubModerators(c *gin.Context) {
	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT u.user_id, u.first_name, u.last_name, u.email, cm.assigned_at
		 FROM club_moderators cm
		 JOIN users u ON cm.user_id = u.user_id
		 WHERE cm.club_id = $1 AND u.is_active = TRUE`,
		clubID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch moderators")
		return
	}
	defer rows.Close()

	type Moderator struct {
		UserID     int    `json:"user_id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Email      string `json:"email"`
		AssignedAt string `json:"assigned_at"`
	}

	var moderators []Moderator

	for rows.Next() {
		var mod Moderator
		var lastNameNS sql.NullString
		var emailNS sql.NullString
		var assignedNT sql.NullTime
		err := rows.Scan(&mod.UserID, &mod.FirstName, &lastNameNS, &emailNS, &assignedNT)
		if err != nil {
			continue
		}
		mod.LastName = models.NullString(lastNameNS)
		mod.Email = models.NullString(emailNS)
		if assignedNT.Valid {
			mod.AssignedAt = assignedNT.Time.Format(time.RFC3339)
		} else {
			mod.AssignedAt = ""
		}
		moderators = append(moderators, mod)
	}

	utils.SuccessResponse(c, http.StatusOK, "Club moderators retrieved", moderators)
}

// ActivateClub sets a club to active (admin only)
func ActivateClub(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.UnauthorizedResponse(c, "User not authenticated")
        return
    }

    role, _ := c.Get("role")
    if role != "system_admin" {
        utils.ForbiddenResponse(c, "Only system admins can activate clubs")
        return
    }

    clubID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "Invalid club ID")
        return
    }

    _, err = database.DB.Exec(
        `UPDATE clubs SET is_active = TRUE, updated_at = CURRENT_TIMESTAMP WHERE club_id = $1`,
        clubID,
    )
    if err != nil {
        utils.InternalServerErrorResponse(c, "Failed to activate club")
        return
    }

    LogActivity(userID.(int), "club_activated", "club", clubID, nil)

    utils.SuccessResponse(c, http.StatusOK, "Club activated successfully", nil)
}

// DeactivateClub sets a club to inactive (admin only)
func DeactivateClub(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        utils.UnauthorizedResponse(c, "User not authenticated")
        return
    }

    role, _ := c.Get("role")
    if role != "system_admin" {
        utils.ForbiddenResponse(c, "Only system admins can deactivate clubs")
        return
    }

    clubID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        utils.BadRequestResponse(c, "Invalid club ID")
        return
    }

    _, err = database.DB.Exec(
        `UPDATE clubs SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE club_id = $1`,
        clubID,
    )
    if err != nil {
        utils.InternalServerErrorResponse(c, "Failed to deactivate club")
        return
    }

    LogActivity(userID.(int), "club_deactivated", "club", clubID, nil)

    utils.SuccessResponse(c, http.StatusOK, "Club deactivated successfully", nil)
}

// AdminGetAllClubs retrieves all clubs (active and inactive) for admin
func AdminGetAllClubs(c *gin.Context) {
    role, _ := c.Get("role")
    if role != "system_admin" {
        utils.ForbiddenResponse(c, "Only system admins can view all clubs")
        return
    }

    rows, err := database.DB.Query(
        `SELECT club_id, club_name, club_code, description, logo_url, cover_image_url, is_active, created_at, updated_at
         FROM clubs
         ORDER BY club_name`,
    )
    if err != nil {
        utils.InternalServerErrorResponse(c, "Failed to fetch clubs")
        return
    }
    defer rows.Close()

    clubs := make([]models.Club, 0)
    for rows.Next() {
        var club models.Club
        var logoURL sql.NullString
        var coverImageURL sql.NullString
        err := rows.Scan(&club.ClubID, &club.ClubName, &club.ClubCode, &club.Description, &logoURL, &coverImageURL, &club.IsActive, &club.CreatedAt, &club.UpdatedAt)
        if err != nil {
            continue
        }
        if logoURL.Valid {
            club.LogoURL = logoURL.String
        }
        if coverImageURL.Valid {
            club.CoverImageURL = coverImageURL.String
        }
        clubs = append(clubs, club)
    }

    utils.SuccessResponse(c, http.StatusOK, "All clubs retrieved", clubs)
}

// Helper function to log activity
func LogActivity(userID int, action, entityType string, entityID int, details interface{}) {
	var detailsJSON string
	if details != nil {
		jsonBytes, _ := json.Marshal(details)
		detailsJSON = string(jsonBytes)
	}

	database.DB.Exec(
		`INSERT INTO activity_log (user_id, action, entity_type, entity_id, details)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, action, entityType, entityID, detailsJSON,
	)
}
