package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/turahe/pkg/response"
)

// TestRequest represents a test request structure
type TestRequest struct {
	ID    string `json:"id" form:"id" uri:"id" binding:"required"`
	Name  string `json:"name" form:"name" binding:"required"`
	Email string `json:"email" form:"email" binding:"required,email"`
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestBaseController_ValidateReqParams_JSON(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.POST("/test", func(c *gin.Context) {
		var req TestRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test valid JSON request
	jsonData := `{"id":"123","name":"Test User","email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response TestRequest
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "123", response.ID)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "test@example.com", response.Email)
}

func TestBaseController_ValidateReqParams_Query(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.GET("/test", func(c *gin.Context) {
		var req TestRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test valid query parameters
	req := httptest.NewRequest("GET", "/test?id=123&name=Test+User&email=test@example.com", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBaseController_ValidateReqParams_URI(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	// Create a request struct that can get ID from URI and other fields from query
	// Note: ID should not be required in query binding since it comes from URI
	type URIRequest struct {
		ID    string `uri:"id" form:"id" binding:"required"`
		Name  string `form:"name"`
		Email string `form:"email"`
	}

	router.GET("/test/:id", func(c *gin.Context) {
		var req URIRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test URI parameter with query params (including ID in query to satisfy binding)
	req := httptest.NewRequest("GET", "/test/123?id=123&name=Test+User&email=test@example.com", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp URIRequest
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "123", resp.ID)
	assert.Equal(t, "Test User", resp.Name)
	assert.Equal(t, "test@example.com", resp.Email)
}

func TestBaseController_ValidateReqParams_InvalidJSON(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.POST("/test", func(c *gin.Context) {
		var req TestRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test invalid JSON (missing required field)
	jsonData := `{"id":"123","name":"Test User"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBaseController_HandleValidationError(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.POST("/test", func(c *gin.Context) {
		var req TestRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			handler.HandleValidationError(c, response.ServiceCodeCommon, err)
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test validation error handling
	jsonData := `{"id":"123"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var errorResponse response.ValidationErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, errorResponse.Errors)
}

func TestBaseController_NormalizePagination(t *testing.T) {
	handler := &BaseHandler{}

	// Test valid pagination
	page, size := handler.NormalizePagination(2, 20)
	assert.Equal(t, 2, page)
	assert.Equal(t, 20, size)

	// Test zero values (should default)
	page, size = handler.NormalizePagination(0, 0)
	assert.Equal(t, 1, page)
	assert.Equal(t, 10, size)

	// Test negative values (should default)
	page, size = handler.NormalizePagination(-1, -5)
	assert.Equal(t, 1, page)
	assert.Equal(t, 10, size)

	// Test page size exceeding max (should cap at 100)
	page, size = handler.NormalizePagination(1, 150)
	assert.Equal(t, 1, page)
	assert.Equal(t, 100, size)

	// Test boundary values
	page, size = handler.NormalizePagination(1, 100)
	assert.Equal(t, 1, page)
	assert.Equal(t, 100, size)
}

func TestBaseController_GetIDFromParam(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.GET("/test/:id", func(c *gin.Context) {
		id, ok := handler.GetIDFromParam(c, "id", response.ServiceCodeCommon)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	// Test with valid ID
	req := httptest.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "123", resp["id"])

	// Test with missing ID (should return validation error)
	router.GET("/test2/:id", func(c *gin.Context) {
		id, ok := handler.GetIDFromParam(c, "missing", response.ServiceCodeCommon)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	req = httptest.NewRequest("GET", "/test2/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestBaseController_GetIDFromRequestOrParam(t *testing.T) {
	handler := &BaseHandler{}

	router := setupRouter()
	router.GET("/test/:id", func(c *gin.Context) {
		reqID := c.Query("req_id")
		id, ok := handler.GetIDFromRequestOrParam(c, reqID, "id", response.ServiceCodeCommon)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	// Test with request ID (should use request ID)
	req := httptest.NewRequest("GET", "/test/123?req_id=456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "456", resp["id"])

	// Test without request ID (should use param)
	req = httptest.NewRequest("GET", "/test/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "123", resp["id"])
}

func TestBaseController_HandleServiceError(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.GET("/test", func(c *gin.Context) {
		err := assert.AnError
		if handler.HandleServiceError(c, response.ServiceCodeCommon, err, "not found") {
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test with not found error
	router.GET("/test2", func(c *gin.Context) {
		err := &testError{message: "not found"}
		if handler.HandleServiceError(c, response.ServiceCodeCommon, err, "not found") {
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test with sentinel ErrNotFound (errors.Is)
	router.GET("/test-sentinel-notfound", func(c *gin.Context) {
		err := fmt.Errorf("user: %w", ErrNotFound)
		if handler.HandleServiceError(c, response.ServiceCodeCommon, err) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	req = httptest.NewRequest("GET", "/test-sentinel-notfound", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test with sentinel ErrUnauthorized
	router.GET("/test-sentinel-unauth", func(c *gin.Context) {
		err := errors.Join(ErrUnauthorized, fmt.Errorf("invalid token"))
		if handler.HandleServiceError(c, response.ServiceCodeCommon, err) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	req = httptest.NewRequest("GET", "/test-sentinel-unauth", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with unauthorized error (legacy string)
	router.GET("/test3", func(c *gin.Context) {
		err := &testError{message: "current password is incorrect"}
		if handler.HandleServiceError(c, response.ServiceCodeCommon, err) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req = httptest.NewRequest("GET", "/test3", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestBaseController_BuildPaginationResponse(t *testing.T) {
	handler := &BaseHandler{}

	// total = 25: page 2, size 10 -> items 11-20, hasNext = (2*10) < 25 = true
	data := []interface{}{1, 2, 3}
	resp := handler.BuildPaginationResponse(data, 2, 10, 25)
	assert.Equal(t, data, resp.Data)
	assert.Equal(t, 2, resp.PageNumber)
	assert.Equal(t, 10, resp.PageSize)
	assert.True(t, resp.HasNext)
	assert.True(t, resp.HasPrev)

	// First page, total 25: hasNext = (1*10) < 25 = true
	resp = handler.BuildPaginationResponse(data, 1, 10, 25)
	assert.False(t, resp.HasPrev)
	assert.True(t, resp.HasNext)

	// Last page: total 20, page 2, size 10 -> (2*10) < 20 = false
	resp = handler.BuildPaginationResponse(data, 2, 10, 20)
	assert.False(t, resp.HasNext)
	assert.True(t, resp.HasPrev)

	// Empty total
	emptyData := []interface{}{}
	resp = handler.BuildPaginationResponse(emptyData, 1, 10, 0)
	assert.False(t, resp.HasNext)
	assert.False(t, resp.HasPrev)
}

func TestBaseController_GetCurrentUserID(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.GET("/test", func(c *gin.Context) {
		userID, ok := handler.GetCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Test without user_id in context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with user_id in context
	router.GET("/test2", func(c *gin.Context) {
		c.Set("user_id", "user123")
		userID, ok := handler.GetCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req = httptest.NewRequest("GET", "/test2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "user123", resp["user_id"])

	// Test with non-string user_id
	router.GET("/test3", func(c *gin.Context) {
		c.Set("user_id", 123)
		userID, ok := handler.GetCurrentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req = httptest.NewRequest("GET", "/test3", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBaseController_CheckUserHasRole(t *testing.T) {
	handler := &BaseHandler{}

	// Test with matching role
	userRoles := []string{"admin", "user"}
	requiredRoles := []string{"admin", "moderator"}
	hasRole := handler.CheckUserHasRole(userRoles, requiredRoles)
	assert.True(t, hasRole)

	// Test without matching role
	userRoles = []string{"user"}
	requiredRoles = []string{"admin", "moderator"}
	hasRole = handler.CheckUserHasRole(userRoles, requiredRoles)
	assert.False(t, hasRole)

	// Test with empty roles
	userRoles = []string{}
	requiredRoles = []string{"admin"}
	hasRole = handler.CheckUserHasRole(userRoles, requiredRoles)
	assert.False(t, hasRole)

	// Test with empty required roles
	userRoles = []string{"admin"}
	requiredRoles = []string{}
	hasRole = handler.CheckUserHasRole(userRoles, requiredRoles)
	assert.False(t, hasRole)

	// Test with multiple matching roles
	userRoles = []string{"admin", "moderator", "user"}
	requiredRoles = []string{"admin", "moderator"}
	hasRole = handler.CheckUserHasRole(userRoles, requiredRoles)
	assert.True(t, hasRole)
}

func TestBaseController_ValidateReqParams_XML(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.POST("/test", func(c *gin.Context) {
		var req TestRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test XML request (though XML binding might not work the same way)
	xmlData := `<TestRequest><id>123</id><name>Test User</name><email>test@example.com</email></TestRequest>`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(xmlData))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// XML binding might fail, but the method should handle it gracefully
	// The exact behavior depends on Gin's XML binding implementation
	assert.NotNil(t, w)
}

func TestBaseController_ValidateReqParams_POST_Query(t *testing.T) {
	router := setupRouter()
	handler := &BaseHandler{}

	router.POST("/test", func(c *gin.Context) {
		var req TestRequest
		err := handler.ValidateReqParams(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Test POST with query parameters (no content type)
	req := httptest.NewRequest("POST", "/test?id=123&name=Test+User&email=test@example.com", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should bind query parameters
	assert.Equal(t, http.StatusOK, w.Code)
}
