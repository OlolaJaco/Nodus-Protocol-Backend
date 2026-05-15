package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API envelope.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents a structured error payload.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// OK sends a 200 JSON response.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

// Created sends a 201 JSON response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

// NoContent sends a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   &APIError{Code: code, Message: message, Details: details},
	})
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error:   &APIError{Code: "UNAUTHORIZED", Message: message},
	})
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error:   &APIError{Code: "FORBIDDEN", Message: message},
	})
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   &APIError{Code: "NOT_FOUND", Message: resource + " not found"},
	})
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error:   &APIError{Code: "CONFLICT", Message: message},
	})
}

// UnprocessableEntity sends a 422 error response.
func UnprocessableEntity(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error:   &APIError{Code: code, Message: message, Details: details},
	})
}

// InternalServerError sends a 500 error response.
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   &APIError{Code: "INTERNAL_ERROR", Message: message},
	})
}

// TooManyRequests sends a 429 error response.
func TooManyRequests(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Error:   &APIError{Code: "RATE_LIMIT_EXCEEDED", Message: "too many requests, please slow down"},
	})
}
