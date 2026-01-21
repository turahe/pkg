package response

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationErrorResponse represents Laravel-style validation error response
type ValidationErrorResponse struct {
	Code    int                 `json:"code"`    // Custom response code
	Message string              `json:"message"` // General error message
	Errors  map[string][]string `json:"errors"`  // Field-specific errors
}

// FormatValidationError formats validation errors in Laravel style
func FormatValidationError(err error) map[string][]string {
	errors := make(map[string][]string)

	// Check if it's a validator.ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := getFieldName(fieldError)
			errorMessage := getErrorMessage(fieldError, fieldName)
			errors[fieldName] = append(errors[fieldName], errorMessage)
		}
	} else {
		// For other types of errors, use a generic message
		errors["general"] = []string{err.Error()}
	}

	return errors
}

// getFieldName extracts the field name from validation error
// Converts struct field names to JSON field names using JSON tags
func getFieldName(fieldError validator.FieldError) string {
	// Get the struct field name (e.g., "Email", "FirstName")
	structField := fieldError.StructField()

	// Get the namespace to extract struct type (e.g., "UserCreateRequest.Email")
	namespace := fieldError.Namespace()

	// Try to extract JSON tag using reflection
	if structField != "" && namespace != "" {
		// Parse namespace to get struct type name
		parts := strings.Split(namespace, ".")
		if len(parts) >= 2 {
			// Try to get the struct type using reflection
			// This requires the struct type to be registered or accessible
			// For now, we'll use a simpler approach: extract from Field() which often contains JSON tag info
			fieldName := fieldError.Field()

			// If Field() returns the JSON tag name directly, use it
			// Otherwise, convert struct field name to camelCase
			if fieldName != structField && fieldName != "" {
				// Field() might already be the JSON tag name
				return fieldName
			}

			// Convert struct field name to camelCase (e.g., "FirstName" -> "firstName")
			return toCamelCase(structField)
		}
	}

	// Fallback: use Field() if available, otherwise convert StructField() to camelCase
	fieldName := fieldError.Field()
	if fieldName != "" {
		return fieldName
	}

	if structField != "" {
		return toCamelCase(structField)
	}

	// Last resort: lowercase
	return strings.ToLower(fieldName)
}

// toCamelCase converts "FirstName" to "firstName"
func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Check if already in camelCase (starts with lowercase)
	if s[0] >= 'a' && s[0] <= 'z' {
		return s
	}

	// Convert first letter to lowercase
	return strings.ToLower(s[:1]) + s[1:]
}

// getErrorMessage generates a human-readable error message from validation error
func getErrorMessage(fieldError validator.FieldError, fieldName string) string {

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required.", fieldName)
	case "email":
		return fmt.Sprintf("The %s must be a valid email address.", fieldName)
	case "min":
		return fmt.Sprintf("The %s must be at least %s characters.", fieldName, fieldError.Param())
	case "max":
		return fmt.Sprintf("The %s may not be greater than %s characters.", fieldName, fieldError.Param())
	case "len":
		return fmt.Sprintf("The %s must be exactly %s characters.", fieldName, fieldError.Param())
	case "numeric":
		return fmt.Sprintf("The %s must be a number.", fieldName)
	case "alpha":
		return fmt.Sprintf("The %s may only contain letters.", fieldName)
	case "alphanum":
		return fmt.Sprintf("The %s may only contain letters and numbers.", fieldName)
	case "url":
		return fmt.Sprintf("The %s must be a valid URL.", fieldName)
	case "uuid":
		return fmt.Sprintf("The %s must be a valid UUID.", fieldName)
	case "oneof":
		return fmt.Sprintf("The %s must be one of: %s.", fieldName, fieldError.Param())
	case "gte":
		return fmt.Sprintf("The %s must be greater than or equal to %s.", fieldName, fieldError.Param())
	case "lte":
		return fmt.Sprintf("The %s must be less than or equal to %s.", fieldName, fieldError.Param())
	case "gt":
		return fmt.Sprintf("The %s must be greater than %s.", fieldName, fieldError.Param())
	case "lt":
		return fmt.Sprintf("The %s must be less than %s.", fieldName, fieldError.Param())
	case "eq":
		return fmt.Sprintf("The %s must be equal to %s.", fieldName, fieldError.Param())
	case "ne":
		return fmt.Sprintf("The %s must not be equal to %s.", fieldName, fieldError.Param())
	case "unique":
		return fmt.Sprintf("The %s has already been taken.", fieldName)
	case "exists":
		return fmt.Sprintf("The selected %s is invalid.", fieldName)
	case "date":
		return fmt.Sprintf("The %s must be a valid date.", fieldName)
	case "datetime":
		return fmt.Sprintf("The %s must be a valid date and time.", fieldName)
	case "timezone":
		return fmt.Sprintf("The %s must be a valid timezone.", fieldName)
	case "json":
		return fmt.Sprintf("The %s must be a valid JSON string.", fieldName)
	case "ip":
		return fmt.Sprintf("The %s must be a valid IP address.", fieldName)
	case "ipv4":
		return fmt.Sprintf("The %s must be a valid IPv4 address.", fieldName)
	case "ipv6":
		return fmt.Sprintf("The %s must be a valid IPv6 address.", fieldName)
	case "base64":
		return fmt.Sprintf("The %s must be a valid base64 string.", fieldName)
	case "required_if":
		return fmt.Sprintf("The %s field is required when %s is present.", fieldName, fieldError.Param())
	case "required_unless":
		return fmt.Sprintf("The %s field is required unless %s is present.", fieldName, fieldError.Param())
	case "required_with":
		return fmt.Sprintf("The %s field is required when %s is present.", fieldName, fieldError.Param())
	case "required_without":
		return fmt.Sprintf("The %s field is required when %s is not present.", fieldName, fieldError.Param())
	default:
		// Generic error message
		if fieldError.Param() != "" {
			return fmt.Sprintf("The %s field is invalid. (%s: %s)", fieldName, fieldError.Tag(), fieldError.Param())
		}
		return fmt.Sprintf("The %s field is invalid. (%s)", fieldName, fieldError.Tag())
	}
}

// GetJSONFieldName extracts JSON tag name from struct field
func GetJSONFieldName(structType reflect.Type, fieldName string) string {
	field, found := structType.FieldByName(fieldName)
	if !found {
		return toCamelCase(fieldName)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return toCamelCase(fieldName)
	}

	// Remove options like "omitempty"
	jsonTag = strings.Split(jsonTag, ",")[0]
	if jsonTag == "" {
		return toCamelCase(fieldName)
	}

	return jsonTag
}
