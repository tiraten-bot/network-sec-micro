package warrior

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorMsg represents a validation error message
type ErrorMsg struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// getErrorMsg returns a map of validator errors
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email"
	case "min":
		return "Must be at least " + fe.Param() + " characters"
	case "max":
		return "Must be at most " + fe.Param() + " characters"
	case "oneof":
		return "Must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "alphanum":
		return "Must contain only letters and numbers"
	case "url":
		return "Must be a valid URL"
	}
	return "Unknown error"
}

// ValidatorMiddleware validates request payloads
func ValidatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware can be extended to validate different types of requests
		c.Next()
	}
}

// ValidateRequest validates the request body
func ValidateRequest(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindJSON(req); err != nil {
		var errors []ErrorMsg
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ErrorMsg{
				Field:   err.Field(),
				Message: getErrorMsg(err),
			})
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"details": errors,
		})
		return false
	}
	return true
}

// ValidateQuery validates query parameters
func ValidateQuery(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindQuery(req); err != nil {
		// Check if it's a validation error
		if ve, ok := err.(validator.ValidationErrors); ok {
			var errors []ErrorMsg
			for _, err := range ve {
				errors = append(errors, ErrorMsg{
					Field:   err.Field(),
					Message: getErrorMsg(err),
				})
			}
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_failed",
				"details": errors,
			})
			return false
		}
		// Other binding errors
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "binding_failed",
			"message": err.Error(),
		})
		return false
	}
	return true
}
