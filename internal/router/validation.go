package router

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResponse represents the response for validation errors
type ValidationResponse struct {
	Errors []ValidationError `json:"errors"`
}

// ValidateJSON validates that the request body is valid JSON
func ValidateJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == "POST" || c.Request().Method == "PUT" || c.Request().Method == "PATCH" {
			if c.Request().Header.Get("Content-Type") == "application/json" {
				var v interface{}
				if err := json.NewDecoder(c.Request().Body).Decode(&v); err != nil {
					return c.JSON(http.StatusBadRequest, ValidationResponse{
						Errors: []ValidationError{
							{Field: "body", Message: "Invalid JSON"},
						},
					})
				}
			}
		}
		return next(c)
	}
}

// ValidateRequiredFields validates that required fields are present
func ValidateRequiredFields(requiredFields []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var body map[string]interface{}
			if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
				return next(c)
			}

			var errors []ValidationError
			for _, field := range requiredFields {
				if _, exists := body[field]; !exists {
					errors = append(errors, ValidationError{
						Field:   field,
						Message: "Field is required",
					})
				}
			}

			if len(errors) > 0 {
				return c.JSON(http.StatusBadRequest, ValidationResponse{Errors: errors})
			}

			return next(c)
		}
	}
}
