package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"real-time-leaderboard/internal/shared/errors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function to use json tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Validate validates a struct
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// formatValidationError formats validation errors
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			message := formatFieldError(fieldError)
			messages = append(messages, message)
		}
		return errors.NewValidationError(strings.Join(messages, "; "), err)
	}
	return errors.NewValidationError("Validation failed", err)
}

// formatFieldError formats a single field error
func formatFieldError(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}
