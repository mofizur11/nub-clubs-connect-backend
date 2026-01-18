package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/handlers"
	"github.com/nub-clubs-connect/nub_admin_api/middleware"
)

func SetupRoutes(router *gin.Engine) {
	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Authentication routes (no auth required)
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/forgot-password", handlers.ForgotPassword)
		authGroup.POST("/reset-password", handlers.ResetPassword)
	}

	// User routes
	userGroup := router.Group("/api/users")
	userGroup.Use(middleware.AuthMiddleware())
	{
		userGroup.GET("/profile", handlers.GetProfile)
		userGroup.PUT("/profile", handlers.UpdateProfile)
		userGroup.GET("/:id/clubs", handlers.GetUserClubs)
		userGroup.GET("/:id/events", handlers.GetUserRegisteredEvents)
	}

	// Club routes
	clubGroup := router.Group("/api/clubs")
	clubGroup.Use(middleware.OptionalAuthMiddleware())
	{
		clubGroup.GET("", handlers.GetAllClubs)
		clubGroup.GET("/:id", handlers.GetClubDetails)
		clubGroup.GET("/:id/members", handlers.GetClubMembers)
		clubGroup.GET("/:id/moderators", handlers.GetClubModerators)
		clubGroup.GET("/:id/news", handlers.GetClubNews)
	}

	// Club routes requiring authentication
	clubAuthGroup := router.Group("/api/clubs")
	clubAuthGroup.Use(middleware.AuthMiddleware())
	{
		clubAuthGroup.POST("", middleware.RoleMiddleware("system_admin"), handlers.CreateClub)
		clubAuthGroup.POST("/:id/join", handlers.JoinClub)
		clubAuthGroup.POST("/:id/leave", handlers.LeaveClub)
		clubAuthGroup.POST("/:id/moderators", middleware.RoleMiddleware("system_admin"), handlers.AssignModerator)
		clubAuthGroup.PUT("/:id/activate", middleware.RoleMiddleware("system_admin"), handlers.ActivateClub)
		clubAuthGroup.PUT("/:id/deactivate", middleware.RoleMiddleware("system_admin"), handlers.DeactivateClub)
		clubAuthGroup.PUT("/:id", middleware.RoleMiddleware("system_admin"), handlers.UpdateClub)
		clubAuthGroup.DELETE("/:id/moderators/:userId", middleware.RoleMiddleware("system_admin"), handlers.RemoveModerator)
	}

	// Event routes
	eventGroup := router.Group("/api/events")
	eventGroup.Use(middleware.OptionalAuthMiddleware())
	{
		eventGroup.GET("", handlers.GetAllEvents)
		eventGroup.GET("/:id", handlers.GetEventDetails)
		eventGroup.GET("/:id/registrations", handlers.GetEventRegistrations)
		eventGroup.GET("/:id/feedback", handlers.GetEventFeedback)
	}

	// Event routes requiring authentication
	eventAuthGroup := router.Group("/api/events")
	eventAuthGroup.Use(middleware.AuthMiddleware())
	{
		eventAuthGroup.POST("", handlers.CreateEvent)
		eventAuthGroup.POST("/:id/register", handlers.RegisterForEvent)
		eventAuthGroup.DELETE("/:id/register", handlers.CancelEventRegistration)
		eventAuthGroup.POST("/:id/feedback", handlers.SubmitEventFeedback)
		eventAuthGroup.POST("/:id/attendance", middleware.RoleMiddleware("club_moderator", "system_admin"), handlers.MarkAttendance)
		eventAuthGroup.POST("/:id/approve", middleware.RoleMiddleware("system_admin"), handlers.ApproveEvent)
		eventAuthGroup.POST("/:id/reject", middleware.RoleMiddleware("system_admin"), handlers.RejectEvent)
		eventAuthGroup.POST("/:id/gallery", handlers.UploadEventGallery)
		eventAuthGroup.DELETE("/:id/gallery/:galleryId", handlers.DeleteGalleryImage)
	}

	// Event gallery routes (public)
	eventGalleryGroup := router.Group("/api/events")
	eventGalleryGroup.Use(middleware.OptionalAuthMiddleware())
	{
		eventGalleryGroup.GET("/:id/gallery", handlers.GetEventGallery)
	}

	// Admin event routes
	adminEventGroup := router.Group("/api/admin/events")
	adminEventGroup.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("system_admin"))
	{
		adminEventGroup.GET("/pending", handlers.GetPendingEvents)
	}

	// News routes
	newsGroup := router.Group("/api/news")
	newsGroup.Use(middleware.OptionalAuthMiddleware())
	{
		newsGroup.GET("", handlers.GetAllNews)
		newsGroup.GET("/featured", handlers.GetFeaturedNews)
		newsGroup.GET("/:id", handlers.GetNewsDetails)
	}

	// News routes requiring authentication
	newsAuthGroup := router.Group("/api/news")
	newsAuthGroup.Use(middleware.AuthMiddleware())
	{
		newsAuthGroup.POST("", handlers.CreateNews)
		newsAuthGroup.POST("/:id/media", handlers.UploadNewsMedia)
		newsAuthGroup.DELETE("/:id/media/:mediaId", handlers.DeleteNewsMedia)
		newsAuthGroup.PUT("/:id/approve", middleware.RoleMiddleware("system_admin"), handlers.ApproveNews)
		newsAuthGroup.PUT("/:id/reject", middleware.RoleMiddleware("system_admin"), handlers.RejectNews)
	}

	// News media routes (public)
	newsMediaGroup := router.Group("/api/news")
	newsMediaGroup.Use(middleware.OptionalAuthMiddleware())
	{
		newsMediaGroup.GET("/:id/media", handlers.GetNewsMedia)
	}

	// Admin news routes
	adminNewsGroup := router.Group("/api/admin/news")
	adminNewsGroup.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("system_admin"))
	{
		adminNewsGroup.GET("/pending", handlers.GetPendingNews)
	}

	// Notification routes
	notificationGroup := router.Group("/api/notifications")
	notificationGroup.Use(middleware.AuthMiddleware())
	{
		notificationGroup.GET("", handlers.GetUserNotifications)
		notificationGroup.GET("/unread-count", handlers.GetUnreadNotificationCount)
		notificationGroup.POST("/:id/read", handlers.MarkNotificationAsRead)
		notificationGroup.POST("/mark-all-read", handlers.MarkAllNotificationsAsRead)
		notificationGroup.DELETE("/:id", handlers.DeleteNotification)
	}

	// System announcements routes
	announcementGroup := router.Group("/api/announcements")
	announcementGroup.Use(middleware.OptionalAuthMiddleware())
	{
		announcementGroup.GET("", handlers.GetSystemAnnouncements)
		announcementGroup.GET("/:id", handlers.GetAnnouncementDetails)
	}

	// System announcements routes requiring admin
	announcementAdminGroup := router.Group("/api/announcements")
	announcementAdminGroup.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("system_admin"))
	{
		announcementAdminGroup.POST("", handlers.CreateSystemAnnouncement)
		announcementAdminGroup.PUT("/:id", handlers.UpdateSystemAnnouncement)
		announcementAdminGroup.DELETE("/:id", handlers.DeleteSystemAnnouncement)
	}

	// Activity log routes
	activityGroup := router.Group("/api/activity")
	activityGroup.Use(middleware.AuthMiddleware())
	{
		activityGroup.GET("/current-user", handlers.GetCurrentUserActivityLog)
		activityGroup.GET("/user/:id", handlers.GetUserActivityLog)
	}

	// Activity log routes (admin only)
	activityAdminGroup := router.Group("/api/activity")
	activityAdminGroup.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("system_admin"))
	{
		activityAdminGroup.GET("/all", handlers.GetAllActivityLogs)
	}

	// Admin routes
	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("system_admin"))
	{
		adminGroup.GET("/dashboard", handlers.GetDashboardStats)
		adminGroup.GET("/analytics/clubs", handlers.GetClubActivityMetrics)
		adminGroup.GET("/analytics/users", handlers.GetUserEngagementStats)
		adminGroup.GET("/analytics/trends", handlers.GetRegistrationTrends)
		adminGroup.GET("/analytics/popular-events", handlers.GetMostPopularEvents)
		adminGroup.GET("/activity", handlers.GetRecentActivity)
		adminGroup.GET("/search/events", handlers.SearchEvents)
		adminGroup.GET("/search/news", handlers.SearchNews)

		// Admin clubs list (active and inactive)
		adminGroup.GET("/clubs", handlers.AdminGetAllClubs)

		// Admin user management
		adminGroup.GET("/users", handlers.AdminListUsers)
		adminGroup.GET("/users/:id", handlers.AdminGetUserByID)
		adminGroup.PUT("/users/:id", handlers.AdminUpdateUser)
		adminGroup.DELETE("/users/:id", handlers.AdminDeleteUser)
		adminGroup.PUT("/users/:id/role", handlers.AdminChangeUserRole)
	}
}
