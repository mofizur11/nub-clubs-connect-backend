package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/models"
)

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message, errorCode string) {
	c.JSON(statusCode, models.APIResponse{
		Success: false,
		Message: message,
		Error:   errorCode,
	})
}

// BadRequestResponse sends a 400 bad request response
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, 400, message, "bad_request")
}

// UnauthorizedResponse sends a 401 unauthorized response
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, 401, message, "unauthorized")
}

// ForbiddenResponse sends a 403 forbidden response
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, 403, message, "forbidden")
}

// NotFoundResponse sends a 404 not found response
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, 404, message, "not_found")
}

// ConflictResponse sends a 409 conflict response
func ConflictResponse(c *gin.Context, message string) {
	ErrorResponse(c, 409, message, "conflict")
}

// InternalServerErrorResponse sends a 500 internal server error response
func InternalServerErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, 500, message, "internal_server_error")
}
