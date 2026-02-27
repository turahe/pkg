package handler

import (
	"errors"
	"net/http"

	"github.com/turahe/pkg/logger"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

// BaseHandler provides shared methods for binding, validation, error handling, and pagination.
// Embed or use as a struct; no fields required.
type BaseHandler struct {
}

// ValidateReqParams binds the request into requestParams based on Content-Type and method.
// For application/json uses ShouldBindJSON; for application/xml uses ShouldBindXML;
// for empty Content-Type binds query and URI params. Returns the first binding error if any.
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

// HandleValidationError writes a 422 response with Laravel-style field errors using response.ValidationError.
func (c *BaseHandler) HandleValidationError(ctx *gin.Context, serviceCode string, err error) {
	response.ValidationError(ctx, serviceCode, err)
}

// NormalizePagination clamps pageNumber and pageSize to response.DefaultPageNumber, response.DefaultPageSize,
// and response.MaxPageSize. Returns the normalized (pageNumber, pageSize).
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

// GetIDFromParam returns the URL parameter value for paramName. If empty, writes a 422 validation
// response and returns ("", false); otherwise returns (id, true).
func (c *BaseHandler) GetIDFromParam(ctx *gin.Context, paramName string, serviceCode string) (string, bool) {
	id := ctx.Param(paramName)
	if id == "" {
		logger.Warnf("Missing %s parameter", paramName)
		response.ValidationErrorSimple(ctx, serviceCode, paramName, "The "+paramName+" field is required.")
		return "", false
	}
	return id, true
}

// GetIDFromRequestOrParam returns reqID if non-empty; otherwise returns the result of GetIDFromParam.
func (c *BaseHandler) GetIDFromRequestOrParam(ctx *gin.Context, reqID string, paramName string, serviceCode string) (string, bool) {
	if reqID != "" {
		return reqID, true
	}
	return c.GetIDFromParam(ctx, paramName, serviceCode)
}

// HandleServiceError maps err to an HTTP response and writes it. Returns true if a response was written.
// Checks errors.Is(ErrNotFound) -> 404, errors.Is(ErrUnauthorized) -> 401, then notFoundMessages exact match -> 404,
// then legacy strings "current password is incorrect" / "Unauthorized" -> 401; otherwise 500.
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

// BuildPaginationResponse builds a SimplePaginationResponse from data and pagination state.
// hasNext is derived from (pageNumber * pageSize) < total; hasPrev is pageNumber > 1.
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

// GetCurrentUserID returns the "user_id" value from the Gin context (set by auth middleware). Second return is false if missing or not a string.
func (c *BaseHandler) GetCurrentUserID(ctx *gin.Context) (string, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// CheckUserHasRole returns true if any element of userRoles equals any element of requiredRoles.
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
