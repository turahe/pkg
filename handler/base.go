package handler

import (
	"errors"
	"net/http"

	"github.com/turahe/pkg/logger"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

type BaseHandler struct {
}

func (c *BaseHandler) ValidateReqParams(ctx *gin.Context, requestParams interface{}) error {
	var err error

	switch ctx.ContentType() {
	case "application/json":
		err = ctx.ShouldBindJSON(requestParams)
	case "application/xml":
		err = ctx.ShouldBindXML(requestParams)
	case "":
		// No Content-Type: bind query params then URI path params (e.g. /:id).
		err = ctx.ShouldBindQuery(requestParams)
		_ = ctx.ShouldBindUri(requestParams)
	default:
		err = ctx.ShouldBind(requestParams)
	}

	if err != nil {
		return err
	}

	return nil
}

// HandleValidationError handles validation errors and returns Laravel-style response
func (c *BaseHandler) HandleValidationError(ctx *gin.Context, serviceCode string, err error) {
	response.ValidationError(ctx, serviceCode, err)
}

// NormalizePagination normalizes pagination parameters with defaults and limits (uses response.DefaultPageNumber, DefaultPageSize, MaxPageSize).
func (c *BaseHandler) NormalizePagination(pageNumber, pageSize int) (int, int) {
	if pageNumber <= 0 {
		pageNumber = response.DefaultPageNumber
	}
	if pageSize <= 0 {
		pageSize = response.DefaultPageSize
	} else if pageSize > response.MaxPageSize {
		pageSize = response.MaxPageSize
	}
	return pageNumber, pageSize
}

// GetIDFromParam extracts ID from URL parameter, validates it, and returns error response if missing
func (c *BaseHandler) GetIDFromParam(ctx *gin.Context, paramName string, serviceCode string) (string, bool) {
	id := ctx.Param(paramName)
	if id == "" {
		logger.Warnf("Missing %s parameter", paramName)
		response.ValidationErrorSimple(ctx, serviceCode, paramName, "The "+paramName+" field is required.")
		return "", false
	}
	return id, true
}

// GetIDFromRequestOrParam extracts ID from request body or URL parameter
func (c *BaseHandler) GetIDFromRequestOrParam(ctx *gin.Context, reqID string, paramName string, serviceCode string) (string, bool) {
	if reqID != "" {
		return reqID, true
	}
	return c.GetIDFromParam(ctx, paramName, serviceCode)
}

// HandleServiceError handles service errors with appropriate HTTP status codes.
// It checks for sentinel errors (ErrNotFound, ErrUnauthorized) via errors.Is, then
// falls back to notFoundMessages and legacy string matching for backward compatibility.
func (c *BaseHandler) HandleServiceError(ctx *gin.Context, serviceCode string, err error, notFoundMessages ...string) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	if errors.Is(err, ErrNotFound) {
		logger.Warnf("Resource not found: %s", errMsg)
		response.NotFoundError(ctx, serviceCode, response.CaseCodeNotFound, errMsg)
		return true
	}
	if errors.Is(err, ErrUnauthorized) {
		logger.Warnf("Unauthorized: %s", errMsg)
		response.UnauthorizedError(ctx, errMsg)
		return true
	}

	for _, msg := range notFoundMessages {
		if errMsg == msg {
			logger.Warnf("Resource not found: %s", errMsg)
			response.NotFoundError(ctx, serviceCode, response.CaseCodeNotFound, errMsg)
			return true
		}
	}

	if errMsg == "current password is incorrect" || errMsg == "Unauthorized" {
		logger.Warnf("Unauthorized: %s", errMsg)
		response.UnauthorizedError(ctx, errMsg)
		return true
	}

	logger.Errorf("Service error: %v", err)
	response.FailWithDetailed(ctx, http.StatusInternalServerError, serviceCode, response.CaseCodeInternalError, nil, errMsg)
	return true
}

// BuildPaginationResponse builds a paginated response from data.
// total is the total number of items; hasNext is true when (pageNumber * pageSize) < total.
func (c *BaseHandler) BuildPaginationResponse(data []interface{}, pageNumber, pageSize int, total int64) response.SimplePaginationResponse {
	hasNext := int64(pageNumber)*int64(pageSize) < total
	hasPrev := pageNumber > 1

	return response.SimplePaginationResponse{
		Data:       data,
		PageNumber: pageNumber,
		PageSize:   pageSize,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// GetCurrentUserID extracts user ID from context (set by auth middleware).
func (c *BaseHandler) GetCurrentUserID(ctx *gin.Context) (string, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// CheckUserHasRole checks if the user has any of the required roles.
func (c *BaseHandler) CheckUserHasRole(userRoles []string, requiredRoles []string) bool {
	for _, roleName := range userRoles {
		for _, requiredRole := range requiredRoles {
			if roleName == requiredRole {
				return true
			}
		}
	}
	return false
}
