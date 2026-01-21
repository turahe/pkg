package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommonResponse struct {
	Code    int         `json:"code"` // Custom response code: HTTP_STATUS + SERVICE_CODE + CASE_CODE (e.g., 2000401)
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Result creates a response with custom code system
func Result(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, data interface{}, message string) {
	responseCode := BuildResponseCode(httpStatus, serviceCode, caseCode)
	ctx.JSON(httpStatus, CommonResponse{
		Code:    responseCode,
		Message: message,
		Data:    data,
	})
}

// ResultWithCode creates a response with explicit response code
func ResultWithCode(ctx *gin.Context, httpStatus int, responseCode int, data interface{}, message string) {
	ctx.JSON(httpStatus, CommonResponse{
		Code:    responseCode,
		Message: message,
		Data:    data,
	})
}

// Ok returns a successful response with default success code
func Ok(ctx *gin.Context) {
	Result(ctx, http.StatusOK, ServiceCodeCommon, CaseCodeSuccess, nil, "success")
}

// OkWithMessage returns a successful response with custom message
func OkWithMessage(ctx *gin.Context, message string) {
	Result(ctx, http.StatusOK, ServiceCodeCommon, CaseCodeSuccess, nil, message)
}

// OkWithData returns a successful response with data
func OkWithData(ctx *gin.Context, data interface{}) {
	Result(ctx, http.StatusOK, ServiceCodeCommon, CaseCodeRetrieved, data, "success")
}

// CursorPaginatedResponse represents a cursor-based pagination response with fields at the top level
type CursorPaginatedResponse struct {
	Code       int         `json:"code"`       // Custom response code
	Message    string      `json:"message"`    // Response message
	Data       interface{} `json:"data"`       // The actual data array
	NextCursor *string     `json:"nextCursor"` // Cursor for the next page (null if no more pages)
	HasNext    bool        `json:"hasNext"`    // Whether there are more items available
}

// CursorPaginated returns a cursor-based paginated response with fields at the top level
func CursorPaginated(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, pagination CursorPaginationResponse, message string) {
	responseCode := BuildResponseCode(httpStatus, serviceCode, caseCode)
	ctx.JSON(httpStatus, CursorPaginatedResponse{
		Code:       responseCode,
		Message:    message,
		Data:       pagination.Data,
		NextCursor: pagination.NextCursor,
		HasNext:    pagination.HasNext,
	})
}

// SimplePaginatedResponse represents a simple pagination response with fields at the top level
type SimplePaginatedResponse struct {
	Code       int         `json:"code"`       // Custom response code
	Message    string      `json:"message"`    // Response message
	Data       interface{} `json:"data"`       // The actual data array
	PageNumber int         `json:"pageNumber"` // Current page number
	PageSize   int         `json:"pageSize"`   // Number of items per page
	HasNext    bool        `json:"hasNext"`    // Whether there is a next page
	HasPrev    bool        `json:"hasPrev"`    // Whether there is a previous page
}

// SimplePaginated returns a simple paginated response with fields at the top level
func SimplePaginated(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, pagination SimplePaginationResponse, message string) {
	responseCode := BuildResponseCode(httpStatus, serviceCode, caseCode)
	ctx.JSON(httpStatus, SimplePaginatedResponse{
		Code:       responseCode,
		Message:    message,
		Data:       pagination.Data,
		PageNumber: pagination.PageNumber,
		PageSize:   pagination.PageSize,
		HasNext:    pagination.HasNext,
		HasPrev:    pagination.HasPrev,
	})
}

// OkWithDetailed returns a response with all parameters
func OkWithDetailed(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, data interface{}, message string) {
	Result(ctx, httpStatus, serviceCode, caseCode, data, message)
}

// Created returns a 201 Created response
func Created(ctx *gin.Context, serviceCode string, data interface{}, message string) {
	if message == "" {
		message = "Resource created successfully"
	}
	Result(ctx, http.StatusCreated, serviceCode, CaseCodeCreated, data, message)
}

// Updated returns a 200 OK response for updates
func Updated(ctx *gin.Context, serviceCode string, data interface{}, message string) {
	if message == "" {
		message = "Resource updated successfully"
	}
	Result(ctx, http.StatusOK, serviceCode, CaseCodeUpdated, data, message)
}

// Deleted returns a 200 OK response for deletions
func Deleted(ctx *gin.Context, serviceCode string, message string) {
	if message == "" {
		message = "Resource deleted successfully"
	}
	Result(ctx, http.StatusOK, serviceCode, CaseCodeDeleted, nil, message)
}

// Fail returns an internal server error response
func Fail(ctx *gin.Context) {
	Result(ctx, http.StatusInternalServerError, ServiceCodeCommon, CaseCodeInternalError, nil, "failure")
}

// FailWithMessage returns an internal server error with custom message
func FailWithMessage(ctx *gin.Context, message string) {
	Result(ctx, http.StatusInternalServerError, ServiceCodeCommon, CaseCodeInternalError, nil, message)
}

// FailWithDetailed returns an error response with all parameters
func FailWithDetailed(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, data interface{}, message string) {
	Result(ctx, httpStatus, serviceCode, caseCode, data, message)
}

// ValidationError returns a 400 Bad Request for validation errors in Laravel style
func ValidationError(ctx *gin.Context, serviceCode string, err error) {
	errors := FormatValidationError(err)
	message := "The given data was invalid."

	responseCode := BuildResponseCode(http.StatusUnprocessableEntity, serviceCode, CaseCodeValidationError)

	ctx.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
		Code:    responseCode,
		Message: message,
		Errors:  errors,
	})
}

// ValidationErrorWithMessage returns a 422 Unprocessable Entity for validation errors with custom message and errors map
func ValidationErrorWithMessage(ctx *gin.Context, serviceCode string, message string, errors map[string][]string) {
	if message == "" {
		message = "The given data was invalid."
	}
	if errors == nil {
		errors = make(map[string][]string)
	}

	responseCode := BuildResponseCode(http.StatusUnprocessableEntity, serviceCode, CaseCodeValidationError)

	ctx.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
		Code:    responseCode,
		Message: message,
		Errors:  errors,
	})
}

// ValidationErrorSimple returns a 422 Unprocessable Entity for simple validation errors (single field)
func ValidationErrorSimple(ctx *gin.Context, serviceCode string, fieldName string, errorMessage string) {
	errors := map[string][]string{
		fieldName: {errorMessage},
	}
	ValidationErrorWithMessage(ctx, serviceCode, "The given data was invalid.", errors)
}

// UnauthorizedError returns a 401 Unauthorized response
func UnauthorizedError(ctx *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	Result(ctx, http.StatusUnauthorized, ServiceCodeAuth, CaseCodeUnauthorized, nil, message)
}

// NotFoundError returns a 404 Not Found response
func NotFoundError(ctx *gin.Context, serviceCode, caseCode string, message string) {
	if message == "" {
		message = "Resource not found"
	}
	if caseCode == "" {
		caseCode = CaseCodeNotFound
	}
	Result(ctx, http.StatusNotFound, serviceCode, caseCode, nil, message)
}

// ConflictError returns a 409 Conflict response
func ConflictError(ctx *gin.Context, serviceCode string, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	Result(ctx, http.StatusConflict, serviceCode, CaseCodeConflict, nil, message)
}

// ForbiddenError returns a 403 Forbidden response
func ForbiddenError(ctx *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	Result(ctx, http.StatusForbidden, ServiceCodeAuth, CaseCodePermissionDenied, nil, message)
}
