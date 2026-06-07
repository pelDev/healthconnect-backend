package domain_errors

import (
	"fmt"
	"net/http"
)

// AppError is a domain-specific error interface
type AppError interface {
	error
	GetHTTPCode() int
	GetCode() string       // Machine-readable error code
	GetTitle() string      // User-friendly title
	GetDetail() string     // Detailed message
	GetMetadata() Metadata // Additional context
	GetCause() error       // Underlying error (for wrapping)
}

// Metadata type for error context
type Metadata map[string]interface{}

// ==================== CONCRETE IMPLEMENTATION ====================

type appError struct {
	HTTPCode int      `json:"-"`
	Code     string   `json:"code"`
	Title    string   `json:"title"`
	Detail   string   `json:"detail"`
	Metadata Metadata `json:"metadata,omitempty"`
	Cause    error    `json:"-"`
}

func (e *appError) Error() string {
	if e.Cause != nil {
		return e.Detail + ": " + e.Cause.Error()
	}
	return e.Detail
}

func (e *appError) GetHTTPCode() int {
	return e.HTTPCode
}

func (e *appError) GetCode() string {
	return e.Code
}

func (e *appError) GetTitle() string {
	return e.Title
}

func (e *appError) GetDetail() string {
	return e.Detail
}

func (e *appError) GetMetadata() Metadata {
	return e.Metadata
}

func (e *appError) GetCause() error {
	return e.Cause
}

// ==================== FACTORY FUNCTIONS ====================

func NewAppError(httpCode int, code, title, detail string) AppError {
	return &appError{
		HTTPCode: httpCode,
		Code:     code,
		Title:    title,
		Detail:   detail,
		Metadata: make(Metadata),
	}
}

func NewAppErrorWithCause(httpCode int, code, title, detail string, cause error) AppError {
	return &appError{
		HTTPCode: httpCode,
		Code:     code,
		Title:    title,
		Detail:   detail,
		Cause:    cause,
		Metadata: make(Metadata),
	}
}

func WrapError(err error, httpCode int, code, title, detail string) AppError {
	if ae, ok := err.(AppError); ok {
		// Already an AppError, return as-is or enhance
		return ae
	}
	return NewAppErrorWithCause(httpCode, code, title, detail, err)
}

// ==================== COMMON DOMAIN ERRORS ====================

// Validation errors
var (
	ErrValidation = NewAppError(
		http.StatusBadRequest,
		"VALIDATION_ERROR",
		"Validation Failed",
		"The request contains invalid parameters",
	)

	ErrRequiredField = func(field string) AppError {
		return NewAppError(
			http.StatusBadRequest,
			"REQUIRED_FIELD",
			"Required Field Missing",
			"The field '"+field+"' is required",
		)
	}

	ErrInvalidEmail = NewAppError(
		http.StatusBadRequest,
		"INVALID_EMAIL",
		"Invalid Email",
		"The provided email address is invalid",
	)

	ErrInvalidUUID = NewAppError(
		http.StatusBadRequest,
		"INVALID_UUID",
		"Invalid Identifier",
		"The provided identifier is invalid",
	)
)

// Authentication & Authorization errors
var (
	ErrInvalidCredentials = NewAppError(
		http.StatusBadRequest,
		"INVALID_CREDENTIALS",
		"Unauthorized",
		"Invalid credentials",
	)

	ErrUnauthorized = NewAppError(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"Unauthorized",
		"Authentication is required to access this resource",
	)

	ErrInvalidToken = NewAppError(
		http.StatusUnauthorized,
		"INVALID_TOKEN",
		"Invalid Token",
		"The provided authentication token is invalid or expired",
	)

	ErrForbidden = NewAppError(
		http.StatusForbidden,
		"FORBIDDEN",
		"Forbidden",
		"You don't have permission to access this resource",
	)

	ErrPermissionDenied = func(resource, action string) AppError {
		return NewAppError(
			http.StatusForbidden,
			"PERMISSION_DENIED",
			"Permission Denied",
			"You don't have permission to "+action+" "+resource,
		)
	}
)

// Resource errors
var (
	ErrNotFound = func(resource string, id interface{}) AppError {
		return NewAppError(
			http.StatusNotFound,
			"NOT_FOUND",
			resource+" Not Found",
			"The requested "+resource+" was not found",
		).(*appError).WithMetadata(Metadata{"id": id})
	}

	ErrConflict = func(resource, field string, value interface{}) AppError {
		return NewAppError(
			http.StatusConflict,
			"CONFLICT",
			resource+" Conflict",
			"A "+resource+" with the same "+field+" already exists",
		).(*appError).WithMetadata(Metadata{field: value})
	}

	ErrAlreadyExists = func(resource string) AppError {
		return NewAppError(
			http.StatusConflict,
			"ALREADY_EXISTS",
			resource+" Exists",
			"The "+resource+" already exists",
		)
	}
)

// Business logic errors
var (
	ErrInvalidTransition = func(from, to string) AppError {
		return NewAppError(
			http.StatusBadRequest,
			"INVALID_TRANSITION",
			"Invalid State Transition",
			"Cannot transition from "+from+" to "+to,
		)
	}

	ErrCircularDependency = NewAppError(
		http.StatusBadRequest,
		"CIRCULAR_DEPENDENCY",
		"Circular Dependency",
		"The operation would create a circular dependency",
	)

	ErrResourceInUse = func(resource string) AppError {
		return NewAppError(
			http.StatusConflict,
			"RESOURCE_IN_USE",
			"Resource In Use",
			"The "+resource+" is currently in use and cannot be modified",
		)
	}
)

// System errors
var (
	ErrInternal = NewAppError(
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"Internal Server Error",
		"An unexpected error occurred",
	)

	ErrInternalWithMessage = func(msg string) AppError {
		return NewAppError(
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Internal Server Error",
			fmt.Sprintf("An unexpected error occurred: %s", msg),
		)
	}

	ErrDatabase = func(cause error) AppError {
		return NewAppErrorWithCause(
			http.StatusInternalServerError,
			"DATABASE_ERROR",
			"Database Error",
			"A database error occurred",
			cause,
		)
	}

	ErrExternalService = func(service string, cause error) AppError {
		return NewAppErrorWithCause(
			http.StatusBadGateway,
			"EXTERNAL_SERVICE_ERROR",
			"External Service Error",
			"Error communicating with "+service+" service",
			cause,
		)
	}
)

// ==================== HELPER METHODS ====================

func (e *appError) WithMetadata(md Metadata) *appError {
	if e.Metadata == nil {
		e.Metadata = make(Metadata)
	}
	for k, v := range md {
		e.Metadata[k] = v
	}
	return e
}

func (e *appError) WithCause(cause error) *appError {
	e.Cause = cause
	return e
}

// ==================== UTILITY FUNCTIONS ====================

// IsAppError checks if an error is a domain AppError
func IsAppError(err error) bool {
	_, ok := err.(AppError)
	return ok
}

// GetAppError returns the AppError if the error is one, nil otherwise
func GetAppError(err error) AppError {
	if ae, ok := err.(AppError); ok {
		return ae
	}
	return nil
}

// GetHTTPCode extracts HTTP status code from error
func GetHTTPCode(err error) int {
	if ae := GetAppError(err); ae != nil {
		return ae.GetHTTPCode()
	}
	return http.StatusInternalServerError
}

// ==================== MIDDLEWARE INTEGRATION ====================

// ErrorResponse is the JSON response structure
type ErrorResponse struct {
	Error struct {
		Code   string   `json:"code"`
		Title  string   `json:"title"`
		Detail string   `json:"detail"`
		Meta   Metadata `json:"metadata,omitempty"`
	} `json:"error"`
}

func NewErrorResponse(err AppError) ErrorResponse {
	var resp ErrorResponse
	resp.Error.Code = err.GetCode()
	resp.Error.Title = err.GetTitle()
	resp.Error.Detail = err.GetDetail()
	resp.Error.Meta = err.GetMetadata()
	return resp
}
