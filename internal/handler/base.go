package controllers

import (
	"net/http"

	"github.com/turahe/pkg/logger"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

type BaseController struct {
}

func (c *BaseController) ValidateReqParams(ctx *gin.Context, requestParams interface{}) error {
	var err error

	switch ctx.ContentType() {
	case "application/json":
		err = ctx.ShouldBindJSON(requestParams)
	case "application/xml":
		err = ctx.ShouldBindXML(requestParams)
	case "":
		// For GET requests, bind query parameters first (most common case)
		// Then try URI parameters for path parameters (e.g., /:id)
		if ctx.Request.Method == "GET" {
			err = ctx.ShouldBindQuery(requestParams)
			// Also try URI binding for path parameters, ignore errors if no URI params
			_ = ctx.ShouldBindUri(requestParams)
		} else {
			// For other methods (POST, PUT, DELETE), try query first, then URI
			err = ctx.ShouldBindQuery(requestParams)
			// Also try URI binding for path parameters, ignore errors if no URI params
			_ = ctx.ShouldBindUri(requestParams)
		}
	default:
		err = ctx.ShouldBind(requestParams)
	}

	if err != nil {
		return err
	}

	return nil
}

// HandleValidationError handles validation errors and returns Laravel-style response
func (c *BaseController) HandleValidationError(ctx *gin.Context, serviceCode string, err error) {
	response.ValidationError(ctx, serviceCode, err)
}

// NormalizePagination normalizes pagination parameters with defaults and limits
func (c *BaseController) NormalizePagination(pageNumber, pageSize int) (int, int) {
	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}
	return pageNumber, pageSize
}

// GetIDFromParam extracts ID from URL parameter, validates it, and returns error response if missing
func (c *BaseController) GetIDFromParam(ctx *gin.Context, paramName string, serviceCode string) (string, bool) {
	id := ctx.Param(paramName)
	if id == "" {
		logger.Warnf("Missing %s parameter", paramName)
		response.ValidationErrorSimple(ctx, serviceCode, paramName, "The "+paramName+" field is required.")
		return "", false
	}
	return id, true
}

// GetIDFromRequestOrParam extracts ID from request body or URL parameter
func (c *BaseController) GetIDFromRequestOrParam(ctx *gin.Context, reqID string, paramName string, serviceCode string) (string, bool) {
	if reqID != "" {
		return reqID, true
	}
	return c.GetIDFromParam(ctx, paramName, serviceCode)
}

// HandleServiceError handles service errors with appropriate HTTP status codes
func (c *BaseController) HandleServiceError(ctx *gin.Context, serviceCode string, err error, notFoundMessages ...string) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Check for not found errors
	for _, msg := range notFoundMessages {
		if errMsg == msg {
			logger.Warnf("Resource not found: %s", errMsg)
			response.NotFoundError(ctx, serviceCode, response.CaseCodeNotFound, errMsg)
			return true
		}
	}

	// Check for unauthorized errors
	if errMsg == "current password is incorrect" || errMsg == "Unauthorized" {
		logger.Warnf("Unauthorized: %s", errMsg)
		response.UnauthorizedError(ctx, errMsg)
		return true
	}

	// Default to internal server error
	logger.Errorf("Service error: %v", err)
	response.FailWithDetailed(ctx, http.StatusInternalServerError, serviceCode, response.CaseCodeInternalError, nil, errMsg)
	return true
}

// BuildPaginationResponse builds a paginated response from data
func (c *BaseController) BuildPaginationResponse(data []interface{}, pageNumber, pageSize int, total int64) response.SimplePaginationResponse {
	hasNext := total > 0
	hasPrev := pageNumber > 1

	return response.SimplePaginationResponse{
		Data:       data,
		PageNumber: pageNumber,
		PageSize:   pageSize,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// GetCurrentAdminID extracts admin ID from context
func (c *BaseController) GetCurrentUserID(ctx *gin.Context) (string, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// CheckAdminHasRole checks if admin has any of the required roles
func (c *BaseController) CheckUserHasRole(userRoles []string, requiredRoles []string) bool {
	for _, roleName := range userRoles {
		for _, requiredRole := range requiredRoles {
			if roleName == requiredRole {
				return true
			}
		}
	}
	return false
}
