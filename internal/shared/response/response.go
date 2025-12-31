package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"real-time-leaderboard/internal/shared/errors"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorInfo represents error information in response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// SuccessWithStatus sends a successful response with custom status code
func SuccessWithStatus(c *gin.Context, status int, data interface{}, message string) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	appErr := errors.AsAppError(err)
	
	c.JSON(appErr.HTTPStatus, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		},
	})
}

// ErrorWithStatus sends an error response with custom status code
func ErrorWithStatus(c *gin.Context, status int, code errors.ErrorCode, message string) {
	c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(code),
			Message: message,
		},
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}, message string) {
	SuccessWithStatus(c, http.StatusCreated, data, message)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

