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

// CreateEvent creates a new event
func CreateEvent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req models.CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	var eventID int
	err := database.DB.QueryRow(
		`INSERT INTO events (club_id, created_by, title, description, event_type, location, start_datetime, end_datetime, registration_deadline, capacity, banner_image_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING event_id`,
		req.ClubID, userID, req.Title, req.Description, req.EventType, req.Location, req.StartDatetime, req.EndDatetime, req.RegistrationDeadline, req.Capacity, req.BannerImageURL,
	).Scan(&eventID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create event")
		return
	}

	// Log activity
	LogActivity(userID.(int), "event_created", "event", eventID, nil)

	response := gin.H{
		"event_id": eventID,
		"title":    req.Title,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Event created successfully", response)
}

// GetAllEvents retrieves all approved events
func GetAllEvents(c *gin.Context) {
	rows, err := database.DB.Query(
		`SELECT 
			e.event_id, e.title, e.description, e.event_type, e.location,
			e.start_datetime, e.end_datetime, e.registration_deadline, e.capacity, e.banner_image_url,
			c.club_name, c.club_code,
			COUNT(er.registration_id) FILTER (WHERE er.registration_status = 'confirmed') as registered_count
		 FROM events e
		 JOIN clubs c ON e.club_id = c.club_id
		 LEFT JOIN event_registrations er ON e.event_id = er.event_id
		 WHERE e.status = 'approved'
		 GROUP BY e.event_id, c.club_id
		 ORDER BY e.start_datetime DESC`,
	)

	if err != nil {
		fmt.Printf("DEBUG: Query error: %v\n", err)
		utils.InternalServerErrorResponse(c, "Failed to fetch events")
		return
	}
	defer rows.Close()

	events := make([]models.Event, 0)

	for rows.Next() {
		var event models.Event
		var bannerImageURL sql.NullString
		err := rows.Scan(&event.EventID, &event.Title, &event.Description, &event.EventType, &event.Location,
			&event.StartDatetime, &event.EndDatetime, &event.RegistrationDeadline, &event.Capacity, &bannerImageURL,
			&event.ClubName, &event.ClubCode, &event.RegisteredCount)
		if err != nil {
			fmt.Printf("DEBUG: Scan error: %v\n", err)
			continue
		}
		if bannerImageURL.Valid {
			event.BannerImageURL = bannerImageURL.String
		}
		events = append(events, event)
	}

	fmt.Printf("DEBUG: Events found: %d\n", len(events))
	utils.SuccessResponse(c, http.StatusOK, "Events retrieved successfully", events)
}

// GetEventDetails retrieves detailed information about an event
func GetEventDetails(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	var event models.Event

	err = database.DB.QueryRow(
		`SELECT 
			e.event_id, e.title, e.description, e.event_type, e.location,
			e.start_datetime, e.end_datetime, e.registration_deadline, e.capacity, e.banner_image_url, e.status,
			c.club_name, c.club_code,
			u.first_name || ' ' || u.last_name as created_by_name,
			COUNT(er.registration_id) FILTER (WHERE er.registration_status = 'confirmed') as confirmed_count,
			COUNT(er.registration_id) FILTER (WHERE er.registration_status = 'waitlist') as waitlist_count,
			AVG(ef.rating) as average_rating,
			COUNT(ef.feedback_id) as feedback_count,
			e.club_id, e.created_by
		 FROM events e
		 JOIN clubs c ON e.club_id = c.club_id
		 JOIN users u ON e.created_by = u.user_id
		 LEFT JOIN event_registrations er ON e.event_id = er.event_id
		 LEFT JOIN event_feedback ef ON e.event_id = ef.event_id
		 WHERE e.event_id = $1
		 GROUP BY e.event_id, c.club_id, u.user_id`,
		eventID,
	).Scan(&event.EventID, &event.Title, &event.Description, &event.EventType, &event.Location,
		&event.StartDatetime, &event.EndDatetime, &event.RegistrationDeadline, &event.Capacity, &event.BannerImageURL, &event.Status,
		&event.ClubName, &event.ClubCode, &event.CreatedByName,
		&event.ConfirmedCount, &event.WaitlistCount, &event.AverageRating, &event.FeedbackCount,
		&event.ClubID, &event.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.NotFoundResponse(c, "Event not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to fetch event details")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Event details retrieved", event)
}

// RegisterForEvent registers a user for an event with waitlist support
func RegisterForEvent(c *gin.Context) {
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

	// Check event capacity and current registrations
	var capacity int
	var currentRegistrations int

	err = database.DB.QueryRow(
		`SELECT 
			e.capacity,
			COUNT(er.registration_id) FILTER (WHERE er.registration_status = 'confirmed')
		 FROM events e
		 LEFT JOIN event_registrations er ON e.event_id = er.event_id
		 WHERE e.event_id = $1
		 GROUP BY e.event_id`,
		eventID,
	).Scan(&capacity, &currentRegistrations)

	if err != nil && err != sql.ErrNoRows {
		utils.InternalServerErrorResponse(c, "Failed to check event capacity")
		return
	}

	// Determine registration status
	registrationStatus := "confirmed"
	if currentRegistrations >= capacity {
		registrationStatus = "waitlist"
	}

	// Insert registration
	var registrationID int
	err = database.DB.QueryRow(
		`INSERT INTO event_registrations (event_id, user_id, registration_status)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (event_id, user_id) DO UPDATE SET registration_status = EXCLUDED.registration_status
		 RETURNING registration_id`,
		eventID, userID, registrationStatus,
	).Scan(&registrationID)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to register for event")
		return
	}

	// Log activity
	LogActivity(userID.(int), "event_registered", "event", eventID, nil)

	response := gin.H{
		"registration_id":     registrationID,
		"registration_status": registrationStatus,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Successfully registered for event", response)
}

// CancelEventRegistration cancels a user's event registration
func CancelEventRegistration(c *gin.Context) {
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

	_, err = database.DB.Exec(
		`UPDATE event_registrations
		 SET registration_status = 'cancelled'
		 WHERE event_id = $1 AND user_id = $2`,
		eventID, userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to cancel registration")
		return
	}

	// Log activity
	LogActivity(userID.(int), "event_registration_cancelled", "event", eventID, nil)

	utils.SuccessResponse(c, http.StatusOK, "Registration cancelled successfully", nil)
}

// GetUserRegisteredEvents retrieves all events a user is registered for
func GetUserRegisteredEvents(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			e.event_id, e.title, e.start_datetime, e.location, e.banner_image_url,
			c.club_name, c.club_code,
			er.registration_status, er.registration_date, er.attendance_marked
		 FROM event_registrations er
		 JOIN events e ON er.event_id = e.event_id
		 JOIN clubs c ON e.club_id = c.club_id
		 WHERE er.user_id = $1 AND er.registration_status != 'cancelled'
		 ORDER BY e.start_datetime DESC`,
		userID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch user events")
		return
	}
	defer rows.Close()

	type UserEvent struct {
		EventID            int    `json:"event_id"`
		Title              string `json:"title"`
		StartDatetime      string `json:"start_datetime"`
		Location           string `json:"location"`
		BannerImageURL     string `json:"banner_image_url"`
		ClubName           string `json:"club_name"`
		ClubCode           string `json:"club_code"`
		RegistrationStatus string `json:"registration_status"`
		RegistrationDate   string `json:"registration_date"`
		AttendanceMarked   bool   `json:"attendance_marked"`
	}

	events := make([]UserEvent, 0)

	for rows.Next() {
		var event UserEvent
		err := rows.Scan(&event.EventID, &event.Title, &event.StartDatetime, &event.Location, &event.BannerImageURL,
			&event.ClubName, &event.ClubCode, &event.RegistrationStatus, &event.RegistrationDate, &event.AttendanceMarked)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	utils.SuccessResponse(c, http.StatusOK, "User events retrieved", events)
}

// GetEventRegistrations retrieves all registrations for an event
func GetEventRegistrations(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			u.user_id, u.student_id, u.first_name, u.last_name, u.email,
			er.registration_status, er.registration_date, er.attendance_marked
		 FROM event_registrations er
		 JOIN users u ON er.user_id = u.user_id
		 WHERE er.event_id = $1
		 ORDER BY er.registration_date`,
		eventID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch registrations")
		return
	}
	defer rows.Close()

	type Registration struct {
		UserID             int    `json:"user_id"`
		StudentID       string `json:"student_id"`
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		Email              string `json:"email"`
		RegistrationStatus string `json:"registration_status"`
		RegistrationDate   string `json:"registration_date"`
		AttendanceMarked   bool   `json:"attendance_marked"`
	}

	registrations := make([]Registration, 0)

	for rows.Next() {
		var reg Registration
		err := rows.Scan(&reg.UserID, &reg.StudentID, &reg.FirstName, &reg.LastName, &reg.Email,
			&reg.RegistrationStatus, &reg.RegistrationDate, &reg.AttendanceMarked)
		if err != nil {
			continue
		}
		registrations = append(registrations, reg)
	}

	utils.SuccessResponse(c, http.StatusOK, "Event registrations retrieved", registrations)
}

// MarkAttendance marks a user's attendance for an event
func MarkAttendance(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
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
		`UPDATE event_registrations
		 SET attendance_marked = TRUE
		 WHERE event_id = $1 AND user_id = $2`,
		eventID, req.UserID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to mark attendance")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Attendance marked successfully", nil)
}

// SubmitEventFeedback submits feedback for an event
func SubmitEventFeedback(c *gin.Context) {
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

	var req models.SubmitFeedbackRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body")
		return
	}

	_, err = database.DB.Exec(
		`INSERT INTO event_feedback (event_id, user_id, rating, comment)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (event_id, user_id) DO UPDATE 
		 SET rating = EXCLUDED.rating, comment = EXCLUDED.comment, submitted_at = CURRENT_TIMESTAMP`,
		eventID, userID, req.Rating, req.Comment,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to submit feedback")
		return
	}

	// Mark feedback as submitted in registration
	database.DB.Exec(
		`UPDATE event_registrations
		 SET feedback_submitted = TRUE
		 WHERE event_id = $1 AND user_id = $2`,
		eventID, userID,
	)

	// Log activity
	LogActivity(userID.(int), "feedback_submitted", "event", eventID, nil)

	utils.SuccessResponse(c, http.StatusCreated, "Feedback submitted successfully", nil)
}

// GetEventFeedback retrieves all feedback for an event
func GetEventFeedback(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			ef.feedback_id, ef.rating, ef.comment, ef.submitted_at,
			u.first_name, u.last_name, u.profile_picture_url
		 FROM event_feedback ef
		 JOIN users u ON ef.user_id = u.user_id
		 WHERE ef.event_id = $1
		 ORDER BY ef.submitted_at DESC`,
		eventID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch feedback")
		return
	}
	defer rows.Close()

	var feedbacks []models.EventFeedback

	for rows.Next() {
		var feedback models.EventFeedback
		err := rows.Scan(&feedback.FeedbackID, &feedback.Rating, &feedback.Comment, &feedback.SubmittedAt,
			&feedback.FirstName, &feedback.LastName, &feedback.ProfilePictureURL)
		if err != nil {
			continue
		}
		feedbacks = append(feedbacks, feedback)
	}

	utils.SuccessResponse(c, http.StatusOK, "Event feedback retrieved", feedbacks)
}

// ApproveEvent approves a pending event
func ApproveEvent(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can approve events")
		return
	}

	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE events
		 SET status = 'approved', updated_at = CURRENT_TIMESTAMP
		 WHERE event_id = $1`,
		eventID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to approve event")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Event approved successfully", nil)
}

// RejectEvent rejects a pending event
func RejectEvent(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can reject events")
		return
	}

	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	_, err = database.DB.Exec(
		`UPDATE events
		 SET status = 'rejected', updated_at = CURRENT_TIMESTAMP
		 WHERE event_id = $1`,
		eventID,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to reject event")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Event rejected successfully", nil)
}

// GetPendingEvents retrieves all pending events for admin approval
func GetPendingEvents(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "system_admin" {
		utils.ForbiddenResponse(c, "Only system admins can view pending events")
		return
	}

	rows, err := database.DB.Query(
		`SELECT 
			e.event_id, e.title, e.start_datetime, e.created_at,
			c.club_name, u.first_name || ' ' || u.last_name as created_by
		 FROM events e
		 JOIN clubs c ON e.club_id = c.club_id
		 JOIN users u ON e.created_by = u.user_id
		 WHERE e.status = 'pending'
		 ORDER BY e.created_at DESC`,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch pending events")
		return
	}
	defer rows.Close()

	type PendingEvent struct {
		EventID       int    `json:"event_id"`
		Title         string `json:"title"`
		StartDatetime string `json:"start_datetime"`
		CreatedAt     string `json:"created_at"`
		ClubName      string `json:"club_name"`
		CreatedBy     string `json:"created_by"`
	}

	var events []PendingEvent

	for rows.Next() {
		var event PendingEvent
		err := rows.Scan(&event.EventID, &event.Title, &event.StartDatetime, &event.CreatedAt, &event.ClubName, &event.CreatedBy)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending events retrieved", events)
}
