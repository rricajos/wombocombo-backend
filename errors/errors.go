package errors

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Details: err.Error()}
}

// Common errors
var (
	ErrBadRequest      = New(fiber.StatusBadRequest, "bad request")
	ErrUnauthorized    = New(fiber.StatusUnauthorized, "unauthorized")
	ErrForbidden       = New(fiber.StatusForbidden, "forbidden")
	ErrNotFound        = New(fiber.StatusNotFound, "not found")
	ErrConflict        = New(fiber.StatusConflict, "conflict")
	ErrTooManyRequests = New(fiber.StatusTooManyRequests, "too many requests")
	ErrInternal        = New(fiber.StatusInternalServerError, "internal server error")
)

func BadRequest(msg string) *AppError {
	return New(fiber.StatusBadRequest, msg)
}

func NotFound(resource string) *AppError {
	return New(fiber.StatusNotFound, fmt.Sprintf("%s not found", resource))
}

func Conflict(msg string) *AppError {
	return New(fiber.StatusConflict, msg)
}

func Forbidden(msg string) *AppError {
	return New(fiber.StatusForbidden, msg)
}

func Internal(err error) *AppError {
	return Wrap(fiber.StatusInternalServerError, "internal server error", err)
}

// SendError writes an AppError as JSON response.
func SendError(c *fiber.Ctx, err *AppError) error {
	return c.Status(err.Code).JSON(err)
}
