package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// CreateNews creates a new news post
func CreateNews(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateNewsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var newsID int
	err := database.DB.QueryRow(
		`INSERT INTO news (club_id, created_by, title, content, category, is_featured)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING news_id`,
		req.ClubID, userID, req.Title, req.Content, req.Category, req.IsFeatured,
	).Scan(&newsID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create news post")
		return
	}

	// Log activity
	LogActivity(userID.(int), "news_created", "news", newsID, nil)

	response := gin.H{
		"news_id": newsID,
		"title": req.Title,
	}

	utils.SuccessResponse(c, http.StatusCreated, "News post created successfully", response)
}

// GetAllNews retrieves all news posts
func GetAllNews(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT 
			n.news_id, n.title, n.content, n.category, n.is_featured, n.published_at,
			c.club_name, c.club_code, c.logo_url,
			u.first_name || ' ' || u.last_name as author
		 FROM news n
		 JOIN clubs c ON n.club_id = c.club_id
		 JOIN users u ON n.created_by = u.user_id
		 WHERE n.status IN ('published')
		 ORDER BY n.published_at DESC
		 LIMIT 20`,
	)

	if err != nil {
		fmt.Printf("DEBUG: Query error: %v\n", err)
		utils.InternalServerErrorResponse(c, "Failed to fetch news")
		return
	}
	defer rows.Close()

	type NewsItem struct {
		NewsID       int    `json:"news_id"`
		Title        string `json:"title"`
		Content      string `json:"content"`
		Category     string `json:"category"`
		IsFeatured   bool   `json:"is_featured"`
		PublishedAt  string `json:"published_at"`
		ClubName     string `json:"club_name"`
		ClubCode     string `json:"club_code"`
		LogoURL      string `json:"logo_url"`
		Author       string `json:"author"`
	}

	newsList := make([]NewsItem, 0)

	for rows.Next() {
		var news NewsItem
		var logoURL sql.NullString
		var publishedAt sql.NullString
		err := rows.Scan(&news.NewsID, &news.Title, &news.Content, &news.Category, &news.IsFeatured, &publishedAt,
			&news.ClubName, &news.ClubCode, &logoURL, &news.Author)
		if err != nil {
			fmt.Printf("DEBUG: Scan error: %v\n", err)
			continue
		}
		if logoURL.Valid {
			news.LogoURL = logoURL.String
		}
		if publishedAt.Valid {
			news.PublishedAt = publishedAt.String
		}
		newsList = append(newsList, news)
	}

	fmt.Printf("DEBUG: News items found: %d\n", len(newsList))
	utils.SuccessResponse(c, http.StatusOK, "News retrieved successfully", newsList)
}

// GetFeaturedNews retrieves featured news for homepage
func GetFeaturedNews(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT 
			n.news_id, n.title, n.content, n.published_at,
			c.club_name, c.club_code,
			(SELECT media_url FROM news_media WHERE news_id = n.news_id AND media_type = 'image' ORDER BY display_order LIMIT 1) as featured_image
		 FROM news n
		 JOIN clubs c ON n.club_id = c.club_id
		 WHERE n.status = 'published' AND n.is_featured = TRUE
		 ORDER BY n.published_at DESC
		 LIMIT 5`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch featured news")
		return
	}
	defer rows.Close()

	type FeaturedNews struct {
		NewsID        int    `json:"news_id"`
		Title         string `json:"title"`
		Content       string `json:"content"`
		PublishedAt   string `json:"published_at"`
		ClubName      string `json:"club_name"`
		ClubCode      string `json:"club_code"`
		FeaturedImage string `json:"featured_image"`
	}

	var newsList []FeaturedNews

	for rows.Next() {
		var news FeaturedNews
		var featuredImage sql.NullString
		err := rows.Scan(&news.NewsID, &news.Title, &news.Content, &news.PublishedAt,
			&news.ClubName, &news.ClubCode, &featuredImage)
		if err != nil {
			continue
		}
		if featuredImage.Valid {
			news.FeaturedImage = featuredImage.String
		}
		newsList = append(newsList, news)
	}

	utils.SuccessResponse(c, http.StatusOK, "Featured news retrieved", newsList)
}

// GetNewsDetails retrieves detailed information about a news post
func GetNewsDetails(c *gin.Context) {
	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid news ID")
		return
	}

	var news models.News

	err = database.DB.QueryRow(
		`SELECT 
			n.news_id, n.title, n.content, n.category, n.is_featured, n.published_at, n.created_at, n.updated_at,
			c.club_name, c.club_code,
			u.first_name || ' ' || u.last_name as author,
			n.club_id, n.created_by
		 FROM news n
		 JOIN clubs c ON n.club_id = c.club_id
		 JOIN users u ON n.created_by = u.user_id
		 WHERE n.news_id = $1 AND n.status = 'published'`,
		newsID,
	).Scan(&news.NewsID, &news.Title, &news.Content, &news.Category, &news.IsFeatured, &news.PublishedAt, &news.CreatedAt, &news.UpdatedAt,
		&news.ClubName, &news.ClubCode, &news.Author, &news.ClubID, &news.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "News post not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to fetch news details")
		return
	}

	// Fetch media
	mediaRows, err := database.DB.Query(
		`SELECT media_id, media_type, media_url, caption, display_order, uploaded_at
		 FROM news_media
		 WHERE news_id = $1
		 ORDER BY display_order`,
		newsID,
	)

	if err == nil {
		defer mediaRows.Close()
		for mediaRows.Next() {
			var media models.NewsMedia
			err := mediaRows.Scan(&media.MediaID, &media.MediaType, &media.MediaURL, &media.Caption, &media.DisplayOrder, &media.UploadedAt)
			if err == nil {
				news.Media = append(news.Media, media)
			}
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "News details retrieved", news)
}

// GetClubNews retrieves all published news for a specific club
func GetClubNews(c *gin.Context) {
	clubID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid club ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			n.news_id, n.title, n.content, n.category, n.published_at,
			u.first_name || ' ' || u.last_name as author
		 FROM news n
		 JOIN users u ON n.created_by = u.user_id
		 WHERE n.club_id = $1 AND n.status = 'published'
		 ORDER BY n.published_at DESC`,
		clubID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch club news")
		return
	}
	defer rows.Close()

	type ClubNewsItem struct {
		NewsID      int    `json:"news_id"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Category    string `json:"category"`
		PublishedAt string `json:"published_at"`
		Author      string `json:"author"`
	}

	var newsList []ClubNewsItem

	for rows.Next() {
		var news ClubNewsItem
		err := rows.Scan(&news.NewsID, &news.Title, &news.Content, &news.Category, &news.PublishedAt, &news.Author)
		if err != nil {
			continue
		}
		newsList = append(newsList, news)
	}

	utils.SuccessResponse(c, http.StatusOK, "Club news retrieved", newsList)
}

// AddNewsMedia adds media to a news post
func AddNewsMedia(c *gin.Context) {
	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid news ID")
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
		`INSERT INTO news_media (news_id, media_type, media_url, caption, display_order)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING media_id`,
		newsID, req.MediaType, req.MediaURL, req.Caption, req.DisplayOrder,
	).Scan(&mediaID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to add media")
		return
	}

	response := gin.H{
		"media_id": mediaID,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Media added successfully", response)
}

// ApproveNews approves a pending news post
func ApproveNews(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can approve news")
		return
	}

	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid news ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE news
		 SET status = 'published', published_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		 WHERE news_id = $1`,
		newsID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to approve news")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "News approved successfully", nil)
}

// RejectNews rejects a pending news post
func RejectNews(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can reject news")
		return
	}

	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid news ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE news
		 SET status = 'rejected', updated_at = CURRENT_TIMESTAMP
		 WHERE news_id = $1`,
		newsID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to reject news")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "News rejected successfully", nil)
}

// GetPendingNews retrieves all pending news for admin approval
func GetPendingNews(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view pending news")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			n.news_id, n.title, n.content, n.created_at,
			c.club_name, u.first_name || ' ' || u.last_name as author
		 FROM news n
		 JOIN clubs c ON n.club_id = c.club_id
		 JOIN users u ON n.created_by = u.user_id
		 WHERE n.status = 'pending'
		 ORDER BY n.created_at DESC`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch pending news")
		return
	}
	defer rows.Close()

	type PendingNews struct {
		NewsID    int    `json:"news_id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		ClubName  string `json:"club_name"`
		Author    string `json:"author"`
	}

	var newsList []PendingNews

	for rows.Next() {
		var news PendingNews
		err := rows.Scan(&news.NewsID, &news.Title, &news.Content, &news.CreatedAt, &news.ClubName, &news.Author)
		if err != nil {
			continue
		}
		newsList = append(newsList, news)
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending news retrieved", newsList)
}
