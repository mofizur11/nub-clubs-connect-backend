package models

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	UserID             int       `json:"user_id"`
	StudentID          string    `json:"student_id"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"-"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Role               string    `json:"role"` // student, club_moderator, system_admin
	Phone              string    `json:"phone"`
	ProfilePictureURL  string    `json:"profile_picture_url"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Club represents a university club
type Club struct {
	ClubID          int       `json:"club_id"`
	ClubName        string    `json:"club_name"`
	ClubCode        string    `json:"club_code"`
	Description     string    `json:"description"`
	LogoURL         string    `json:"logo_url"`
	CoverImageURL   string    `json:"cover_image_url"`
	FoundedDate     *time.Time `json:"founded_date"`
	Email           string    `json:"email"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	MemberCount     int       `json:"member_count,omitempty"`
	UpcomingEvents  int       `json:"upcoming_events,omitempty"`
}

// ClubMember represents a user's membership in a club
type ClubMember struct {
	MembershipID int       `json:"membership_id"`
	UserID       int       `json:"user_id"`
	ClubID       int       `json:"club_id"`
	Role         string    `json:"role"` // member, president, vice_president, treasurer, etc.
	JoinedDate   time.Time `json:"joined_date"`
	IsActive     bool      `json:"is_active"`
}

// ClubModerator represents a moderator for a club
type ClubModerator struct {
	ModeratorID int       `json:"moderator_id"`
	UserID      int       `json:"user_id"`
	ClubID      int       `json:"club_id"`
	AssignedAt  time.Time `json:"assigned_at"`
}

// Event represents a club event
type Event struct {
	EventID              int       `json:"event_id"`
	ClubID              int       `json:"club_id"`
	CreatedBy           int       `json:"created_by"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	EventType           string    `json:"event_type"` // workshop, seminar, social, competition, etc.
	Location            string    `json:"location"`
	StartDatetime       time.Time `json:"start_datetime"`
	EndDatetime         time.Time `json:"end_datetime"`
	RegistrationDeadline *time.Time `json:"registration_deadline"`
	Capacity            int       `json:"capacity"`
	IsRegistrationOpen  bool      `json:"is_registration_open"`
	Status              string    `json:"status"` // pending, approved, rejected, completed, cancelled
	BannerImageURL      string    `json:"banner_image_url"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	ClubName            string    `json:"club_name,omitempty"`
	ClubCode            string    `json:"club_code,omitempty"`
	CreatedByName       string    `json:"created_by_name,omitempty"`
	RegisteredCount     int       `json:"registered_count,omitempty"`
	ConfirmedCount      int       `json:"confirmed_count,omitempty"`
	WaitlistCount       int       `json:"waitlist_count,omitempty"`
	AverageRating       *float64  `json:"average_rating,omitempty"`
	FeedbackCount       int       `json:"feedback_count,omitempty"`
}

// EventRegistration represents a user's registration for an event
type EventRegistration struct {
	RegistrationID   int       `json:"registration_id"`
	EventID          int       `json:"event_id"`
	UserID           int       `json:"user_id"`
	RegistrationStatus string  `json:"registration_status"` // confirmed, waitlist, cancelled, attended
	RegistrationDate time.Time `json:"registration_date"`
	AttendanceMarked bool      `json:"attendance_marked"`
	FeedbackSubmitted bool      `json:"feedback_submitted"`
}

// EventFeedback represents feedback for an event
type EventFeedback struct {
	FeedbackID int       `json:"feedback_id"`
	EventID    int       `json:"event_id"`
	UserID     int       `json:"user_id"`
	Rating     int       `json:"rating"` // 1-5
	Comment    string    `json:"comment"`
	SubmittedAt time.Time `json:"submitted_at"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	ProfilePictureURL string `json:"profile_picture_url,omitempty"`
}

// EventGallery represents photos from an event
type EventGallery struct {
	GalleryID  int       `json:"gallery_id"`
	EventID    int       `json:"event_id"`
	UploadedBy int       `json:"uploaded_by"`
	ImageURL   string    `json:"image_url"`
	Caption    string    `json:"caption"`
	UploadedAt time.Time `json:"uploaded_at"`
	UploadedByName string `json:"uploaded_by_name,omitempty"`
}

// News represents a news/announcement post
type News struct {
	NewsID    int       `json:"news_id"`
	ClubID    int       `json:"club_id"`
	CreatedBy int       `json:"created_by"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"` // achievement, announcement, update, etc.
	IsFeatured bool      `json:"is_featured"`
	Status    string    `json:"status"` // pending, rejected, published
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ClubName  string    `json:"club_name,omitempty"`
	ClubCode  string    `json:"club_code,omitempty"`
	Author    string    `json:"author,omitempty"`
	Media     []NewsMedia `json:"media,omitempty"`
}

// NewsMedia represents media attached to a news post
type NewsMedia struct {
	MediaID      int       `json:"media_id"`
	NewsID       int       `json:"news_id"`
	MediaType    string    `json:"media_type"` // image, video
	MediaURL     string    `json:"media_url"`
	Caption      string    `json:"caption"`
	DisplayOrder int       `json:"display_order"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// Notification represents a system notification
type Notification struct {
	NotificationID   int       `json:"notification_id"`
	UserID           int       `json:"user_id"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	NotificationType string    `json:"notification_type"` // event_reminder, news_update, registration_confirmation, etc.
	RelatedEntityType string    `json:"related_entity_type"` // event, news, club
	RelatedEntityID  int       `json:"related_entity_id"`
	IsRead           bool      `json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}

// SystemAnnouncement represents a system-wide announcement
type SystemAnnouncement struct {
	AnnouncementID int       `json:"announcement_id"`
	CreatedBy      int       `json:"created_by"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Priority       string    `json:"priority"` // low, normal, high, urgent
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedByName  string    `json:"created_by_name,omitempty"`
}

// ActivityLog represents a user activity log entry
type ActivityLog struct {
	LogID      int       `json:"log_id"`
	UserID     int       `json:"user_id"`
	Action     string    `json:"action"` // login, event_registration, news_published, etc.
	EntityType string    `json:"entity_type"` // event, news, club, user
	EntityID   int       `json:"entity_id"`
	Details    string    `json:"details"` // JSON string
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
	UserName   string    `json:"user_name,omitempty"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	TokenID   int       `json:"token_id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IsUsed    bool      `json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}

// DashboardStats represents overall statistics
type DashboardStats struct {
	TotalStudents      int `json:"total_students"`
	TotalClubs         int `json:"total_clubs"`
	TotalEvents        int `json:"total_events"`
	TotalRegistrations int `json:"total_registrations"`
	TotalNews          int `json:"total_news"`
}

// ClubActivityMetrics represents activity metrics for a club
type ClubActivityMetrics struct {
	ClubID             int    `json:"club_id"`
	ClubName           string `json:"club_name"`
	MemberCount        int    `json:"member_count"`
	TotalEvents        int    `json:"total_events"`
	TotalRegistrations int    `json:"total_registrations"`
	TotalNews          int    `json:"total_news"`
}

// UserEngagementStats represents user engagement statistics
type UserEngagementStats struct {
	UserID        int    `json:"user_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	ClubsJoined   int    `json:"clubs_joined"`
	EventsAttended int    `json:"events_attended"`
	FeedbackGiven int    `json:"feedback_given"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Phone             string `json:"phone"`
	ProfilePictureURL string `json:"profile_picture_url"`
}

// CreateClubRequest represents a club creation request
type CreateClubRequest struct {
	ClubName      string `json:"club_name" binding:"required"`
	ClubCode      string `json:"club_code" binding:"required"`
	Description   string `json:"description"`
	LogoURL       string `json:"logo_url"`
	CoverImageURL string `json:"cover_image_url"`
	FoundedDate   *time.Time `json:"founded_date"`
	Email         string `json:"email"`
}

// CreateEventRequest represents an event creation request
type CreateEventRequest struct {
	ClubID               int       `json:"club_id" binding:"required"`
	Title                string    `json:"title" binding:"required"`
	Description          string    `json:"description"`
	EventType            string    `json:"event_type"`
	Location             string    `json:"location"`
	StartDatetime        time.Time `json:"start_datetime" binding:"required"`
	EndDatetime          time.Time `json:"end_datetime" binding:"required"`
	RegistrationDeadline *time.Time `json:"registration_deadline"`
	Capacity             int       `json:"capacity"`
	BannerImageURL       string    `json:"banner_image_url"`
}

// CreateNewsRequest represents a news creation request
type CreateNewsRequest struct {
	ClubID     int    `json:"club_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Category   string `json:"category"`
	IsFeatured bool   `json:"is_featured"`
}

// SubmitFeedbackRequest represents feedback submission
type SubmitFeedbackRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// Helper function to check if a value is null
func NullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func NullInt64(i sql.NullInt64) int {
	if i.Valid {
		return int(i.Int64)
	}
	return 0
}

func NullTime(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func NullFloat64(f sql.NullFloat64) *float64 {
	if f.Valid {
		return &f.Float64
	}
	return nil
}
