// Package response provides HTTP response utilities for gin handlers.
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/jaochai/ugc/pkg/errors"
)

// Response represents a standard API response.
type Response struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Meta    *Meta          `json:"meta,omitempty"`
}

// ErrorResponse represents an error in the API response.
type ErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta represents pagination metadata.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta creates a new Meta with calculated TotalPages.
func NewMeta(page, perPage int, total int64) *Meta {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Success sends a successful response with HTTP 200 OK.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with pagination metadata.
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a successful response with HTTP 201 Created.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent sends an empty response with HTTP 204 No Content.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response. It handles AppError specially to extract
// the status code, message, and details. For other errors, it returns
// HTTP 500 Internal Server Error.
func Error(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, Response{
			Success: false,
			Error: &ErrorResponse{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	// For non-AppError errors, return 500 Internal Server Error
	// Don't expose the actual error message for security
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		},
	})
}

// ValidationError sends a validation error response with HTTP 400 Bad Request.
func ValidationError(c *gin.Context, details map[string]string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "validation failed",
			Details: details,
		},
	})
}

// BadRequest sends a bad request error response with HTTP 400.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: message,
		},
	})
}

// Unauthorized sends an unauthorized error response with HTTP 401.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: message,
		},
	})
}

// Forbidden sends a forbidden error response with HTTP 403.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusForbidden,
			Message: message,
		},
	})
}

// NotFound sends a not found error response with HTTP 404.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusNotFound,
			Message: message,
		},
	})
}

// InternalServerError sends an internal server error response with HTTP 500.
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: message,
		},
	})
}
