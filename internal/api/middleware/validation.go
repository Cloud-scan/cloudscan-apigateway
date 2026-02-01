package middleware

import (
	"net/http"

	"github.com/cloud-scan/cloudscan-apigateway/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new custom validator
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates structs
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Return validation errors in a user-friendly format
		validationErrors := err.(validator.ValidationErrors)
		errors := make(map[string]string)
		for _, fieldError := range validationErrors {
			errors[fieldError.Field()] = getErrorMessage(fieldError)
		}
		return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"message": "validation failed",
			"errors":  errors,
		})
	}
	return nil
}

// getErrorMessage returns a user-friendly error message
func getErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "value is too short"
	case "max":
		return "value is too long"
	case "url":
		return "invalid URL format"
	default:
		return "invalid value"
	}
}

// BindAndValidate binds and validates request
func BindAndValidate(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	validator := utils.NewValidator()
	if err := validator.Validate(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}