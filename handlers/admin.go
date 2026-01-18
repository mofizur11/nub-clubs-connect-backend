package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/models"
	"github.com/nub-clubs-connect/nub_admin_api/utils"
)

// GetDashboardStats retrieves overall statistics for the admin dashboard
func GetDashboardStats(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view dashboard")
		return
	}

	var stats models.DashboardStats

	err := database.DB.QueryRow(
		`SELECT 
			(SELECT COUNT(*) FROM users WHERE role = 'student' AND is_active = TRUE) as total_students,
			(SELECT COUNT(*) FROM clubs WHERE is_active = TRUE) as total_clubs,
			(SELECT COUNT(*) FROM events WHERE status = 'approved') as total_events,
			(SELECT COUNT(*) FROM event_registrations WHERE registration_status = 'confirmed') as total_registrations,
			(SELECT COUNT(*) FROM news WHERE status = 'published') as total_news`,
	).Scan(&stats.TotalStudents, &stats.TotalClubs, &stats.TotalEvents, &stats.TotalRegistrations, &stats.TotalNews)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch dashboard statistics")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Dashboard statistics retrieved", stats)
}

// GetClubActivityMetrics retrieves activity metrics for all clubs
func GetClubActivityMetrics(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view analytics")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			c.club_id, c.club_name,
			COUNT(DISTINCT cm.user_id) as member_count,
			COUNT(DISTINCT e.event_id) as total_events,
			COUNT(DISTINCT er.user_id) as total_registrations,
			COUNT(DISTINCT n.news_id) as total_news
		 FROM clubs c
		 LEFT JOIN club_members cm ON c.club_id = cm.club_id AND cm.is_active = TRUE
		 LEFT JOIN events e ON c.club_id = e.club_id AND e.status = 'approved'
		 LEFT JOIN event_registrations er ON e.event_id = er.event_id
		 LEFT JOIN news n ON c.club_id = n.club_id AND n.status = 'published'
		 WHERE c.is_active = TRUE
		 GROUP BY c.club_id
		 ORDER BY total_events DESC`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch club metrics")
		return
	}
	defer rows.Close()

	metrics := make([]models.ClubActivityMetrics, 0)

	for rows.Next() {
		var metric models.ClubActivityMetrics
		err := rows.Scan(&metric.ClubID, &metric.ClubName, &metric.MemberCount, &metric.TotalEvents, &metric.TotalRegistrations, &metric.TotalNews)
		if err != nil {
			continue
		}
		metrics = append(metrics, metric)
	}

	utils.SuccessResponse(c, http.StatusOK, "Club activity metrics retrieved", metrics)
}

// GetUserEngagementStats retrieves user engagement statistics
func GetUserEngagementStats(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view analytics")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			u.user_id, u.first_name, u.last_name, u.email,
			COUNT(DISTINCT cm.club_id) as clubs_joined,
			COUNT(DISTINCT er.event_id) as events_attended,
			COUNT(DISTINCT ef.feedback_id) as feedback_given
		 FROM users u
		 LEFT JOIN club_members cm ON u.user_id = cm.user_id AND cm.is_active = TRUE
		 LEFT JOIN event_registrations er ON u.user_id = er.user_id AND er.attendance_marked = TRUE
		 LEFT JOIN event_feedback ef ON u.user_id = ef.user_id
		 WHERE u.role = 'student'
		 GROUP BY u.user_id
		 ORDER BY events_attended DESC
		 LIMIT 20`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch user engagement stats")
		return
	}
	defer rows.Close()

	stats := make([]models.UserEngagementStats, 0)

	for rows.Next() {
		var stat models.UserEngagementStats
		err := rows.Scan(&stat.UserID, &stat.FirstName, &stat.LastName, &stat.Email, &stat.ClubsJoined, &stat.EventsAttended, &stat.FeedbackGiven)
		if err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	utils.SuccessResponse(c, http.StatusOK, "User engagement statistics retrieved", stats)
}

// GetRegistrationTrends retrieves registration trends for the last 6 months
func GetRegistrationTrends(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view analytics")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			DATE_TRUNC('month', er.registration_date) as month,
			COUNT(*) as registrations
		 FROM event_registrations er
		 WHERE er.registration_date >= CURRENT_DATE - INTERVAL '6 months'
		 GROUP BY month
		 ORDER BY month`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch registration trends")
		return
	}
	defer rows.Close()

	type TrendData struct {
		Month         string `json:"month"`
		Registrations int    `json:"registrations"`
	}

	trends := make([]TrendData, 0)

	for rows.Next() {
		var trend TrendData
		var month sql.NullTime
		err := rows.Scan(&month, &trend.Registrations)
		if err != nil {
			continue
		}
		if month.Valid {
			trend.Month = month.Time.Format("2006-01")
		}
		trends = append(trends, trend)
	}

	utils.SuccessResponse(c, http.StatusOK, "Registration trends retrieved", trends)
}

// GetMostPopularEvents retrieves the most popular events
func GetMostPopularEvents(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT 
			e.event_id, e.title, c.club_name,
			COUNT(er.registration_id) as registration_count,
			AVG(ef.rating) as average_rating
		 FROM events e
		 JOIN clubs c ON e.club_id = c.club_id
		 LEFT JOIN event_registrations er ON e.event_id = er.event_id
		 LEFT JOIN event_feedback ef ON e.event_id = ef.event_id
		 WHERE e.status = 'approved'
		 GROUP BY e.event_id, c.club_id
		 ORDER BY registration_count DESC
		 LIMIT 10`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch popular events")
		return
	}
	defer rows.Close()

	type PopularEvent struct {
		EventID            int     `json:"event_id"`
		Title              string  `json:"title"`
		ClubName           string  `json:"club_name"`
		RegistrationCount  int     `json:"registration_count"`
		AverageRating      *float64 `json:"average_rating"`
	}

	events := make([]PopularEvent, 0)

	for rows.Next() {
		var event PopularEvent
		var avgRating sql.NullFloat64
		err := rows.Scan(&event.EventID, &event.Title, &event.ClubName, &event.RegistrationCount, &avgRating)
		if err != nil {
			continue
		}
		if avgRating.Valid {
			event.AverageRating = &avgRating.Float64
		}
		events = append(events, event)
	}

	utils.SuccessResponse(c, http.StatusOK, "Popular events retrieved", events)
}

// GetRecentActivity retrieves recent activity logs
func GetRecentActivity(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view activity logs")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			al.log_id, al.action, al.entity_type, al.created_at,
			u.first_name || ' ' || u.last_name as user_name,
			al.details
		 FROM activity_log al
		 JOIN users u ON al.user_id = u.user_id
		 ORDER BY al.created_at DESC
		 LIMIT 50`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch activity logs")
		return
	}
	defer rows.Close()

	type ActivityEntry struct {
		LogID      int    `json:"log_id"`
		Action     string `json:"action"`
		EntityType string `json:"entity_type"`
		CreatedAt  string `json:"created_at"`
		UserName   string `json:"user_name"`
		Details    string `json:"details"`
	}

	activities := make([]ActivityEntry, 0)

	for rows.Next() {
		var activity ActivityEntry
		var details sql.NullString
		err := rows.Scan(&activity.LogID, &activity.Action, &activity.EntityType, &activity.CreatedAt, &activity.UserName, &details)
		if err != nil {
			continue
		}
		if details.Valid {
			activity.Details = details.String
		}
		activities = append(activities, activity)
	}

	utils.SuccessResponse(c, http.StatusOK, "Recent activity retrieved", activities)
}

// SearchEvents searches for events by keyword
func SearchEvents(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		utils.BadRequestResponse(c, "Search keyword is required")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			e.event_id, e.title, e.description, e.start_datetime,
			c.club_name, c.club_code
		 FROM events e
		 JOIN clubs c ON e.club_id = c.club_id
		 WHERE e.status = 'approved' 
		     AND (
		         e.title ILIKE '%' || $1 || '%' 
		         OR e.description ILIKE '%' || $1 || '%'
		     )
		 ORDER BY e.start_datetime`,
		keyword,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search events")
		return
	}
	defer rows.Close()

	type SearchResult struct {
		EventID       int    `json:"event_id"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		StartDatetime string `json:"start_datetime"`
		ClubName      string `json:"club_name"`
		ClubCode      string `json:"club_code"`
	}

	results := make([]SearchResult, 0)

	for rows.Next() {
		var result SearchResult
		err := rows.Scan(&result.EventID, &result.Title, &result.Description, &result.StartDatetime, &result.ClubName, &result.ClubCode)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	utils.SuccessResponse(c, http.StatusOK, "Search results retrieved", results)
}

// SearchNews searches for news by keyword
func SearchNews(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		utils.BadRequestResponse(c, "Search keyword is required")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			n.news_id, n.title, n.content, n.published_at,
			c.club_name
		 FROM news n
		 JOIN clubs c ON n.club_id = c.club_id
		 WHERE n.status = 'published'
		     AND (
		         n.title ILIKE '%' || $1 || '%' 
		         OR n.content ILIKE '%' || $1 || '%'
		     )
		 ORDER BY n.published_at DESC`,
		keyword,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to search news")
		return
	}
	defer rows.Close()

	type SearchResult struct {
		NewsID      int    `json:"news_id"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		PublishedAt string `json:"published_at"`
		ClubName    string `json:"club_name"`
	}

	results := make([]SearchResult, 0)

	for rows.Next() {
		var result SearchResult
		err := rows.Scan(&result.NewsID, &result.Title, &result.Content, &result.PublishedAt, &result.ClubName)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	utils.SuccessResponse(c, http.StatusOK, "Search results retrieved", results)
}
