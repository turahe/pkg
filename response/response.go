package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CommonResponse is the standard JSON body: code (7-digit composite), message, data.
type CommonResponse struct {
	Code    int         `json:"code"` // Custom response code: HTTP_STATUS + SERVICE_CODE + CASE_CODE (e.g., 2000401)
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// WriteParams groups fields for a standard JSON response write.
type WriteParams struct {
	HTTPStatus  int
	ServiceCode string
	CaseCode    string
	Data        interface{}
	Message     string
}

// Write writes a JSON response using WriteParams (preferred over Result for call sites with many args).
func Write(ctx *gin.Context, p WriteParams) {
	responseCode := BuildResponseCode(p.HTTPStatus, p.ServiceCode, p.CaseCode)
	ctx.JSON(p.HTTPStatus, CommonResponse{
		Code:    responseCode,
		Message: p.Message,
		Data:    p.Data,
	})
}

// Result writes a JSON response with the given HTTP status and composite code.
// Prefer Write with WriteParams for new call sites.
func Result(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, data interface{}, message string) {
	Write(ctx, WriteParams{
		HTTPStatus:  httpStatus,
		ServiceCode: serviceCode,
		CaseCode:    caseCode,
		Data:        data,
		Message:     message,
	})
}

// ResultWithCode writes a JSON response with explicit responseCode (no BuildResponseCode).
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

// CursorPaginatedResponse is the top-level JSON shape for cursor-based pagination (code, message, data, nextCursor, hasNext).
type CursorPaginatedResponse struct {
	Code       int         `json:"code"`       // Custom response code
	Message    string      `json:"message"`    // Response message
	Data       interface{} `json:"data"`       // The actual data array
	NextCursor *string     `json:"nextCursor"` // Cursor for the next page (null if no more pages)
	HasNext    bool        `json:"hasNext"`    // Whether there are more items available
}

// CursorPaginated writes a cursor-based paginated JSON response using the given pagination payload.
// CursorPaginatedParams groups fields for a cursor-paginated response.
type CursorPaginatedParams struct {
	HTTPStatus  int
	ServiceCode string
	CaseCode    string
	Pagination  CursorPaginationResponse
	Message     string
}

// WriteCursorPaginated writes a cursor-based paginated JSON response.
func WriteCursorPaginated(ctx *gin.Context, p CursorPaginatedParams) {
	responseCode := BuildResponseCode(p.HTTPStatus, p.ServiceCode, p.CaseCode)
	ctx.JSON(p.HTTPStatus, CursorPaginatedResponse{
		Code:       responseCode,
		Message:    p.Message,
		Data:       p.Pagination.Data,
		NextCursor: p.Pagination.NextCursor,
		HasNext:    p.Pagination.HasNext,
	})
}

// CursorPaginated writes a cursor-based paginated JSON response. Prefer WriteCursorPaginated.
func CursorPaginated(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, pagination CursorPaginationResponse, message string) {
	WriteCursorPaginated(ctx, CursorPaginatedParams{
		HTTPStatus:  httpStatus,
		ServiceCode: serviceCode,
		CaseCode:    caseCode,
		Pagination:  pagination,
		Message:     message,
	})
}

// SimplePaginatedResponse is the top-level JSON shape for offset pagination (code, message, data, pageNumber, pageSize, hasNext, hasPrev).
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
// SimplePaginatedParams groups fields for an offset-paginated response.
type SimplePaginatedParams struct {
	HTTPStatus  int
	ServiceCode string
	CaseCode    string
	Pagination  SimplePaginationResponse
	Message     string
}

// WriteSimplePaginated writes an offset-paginated JSON response.
func WriteSimplePaginated(ctx *gin.Context, p SimplePaginatedParams) {
	responseCode := BuildResponseCode(p.HTTPStatus, p.ServiceCode, p.CaseCode)
	ctx.JSON(p.HTTPStatus, SimplePaginatedResponse{
		Code:       responseCode,
		Message:    p.Message,
		Data:       p.Pagination.Data,
		PageNumber: p.Pagination.PageNumber,
		PageSize:   p.Pagination.PageSize,
		HasNext:    p.Pagination.HasNext,
		HasPrev:    p.Pagination.HasPrev,
	})
}

// SimplePaginated writes an offset-paginated JSON response. Prefer WriteSimplePaginated.
func SimplePaginated(ctx *gin.Context, httpStatus int, serviceCode, caseCode string, pagination SimplePaginationResponse, message string) {
	WriteSimplePaginated(ctx, SimplePaginatedParams{
		HTTPStatus:  httpStatus,
		ServiceCode: serviceCode,
		CaseCode:    caseCode,
		Pagination:  pagination,
		Message:     message,
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
