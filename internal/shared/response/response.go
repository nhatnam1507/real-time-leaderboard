// Package response provides HTTP response utilities and error handling.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorInfo represents error information in response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// NewPagination creates pagination metadata from offset, limit, and total count
func NewPagination(offset, limit int, total int64) Pagination {
	page := 1
	if limit > 0 {
		page = (offset / limit) + 1
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
		if totalPages == 0 && total > 0 {
			totalPages = 1
		}
	}

	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
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
	appErr := AsAppError(err)

	c.JSON(appErr.HTTPStatus, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		},
	})
}

// ErrorWithStatus sends an error response with custom status code
func ErrorWithStatus(c *gin.Context, status int, code ErrorCode, message string) {
	c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(code),
			Message: message,
		},
	})
}

// SuccessWithMeta sends a successful response with custom metadata
func SuccessWithMeta(c *gin.Context, data interface{}, message string, meta interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Message: message,
		Meta:    meta,
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
