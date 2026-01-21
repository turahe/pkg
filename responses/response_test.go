package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func testContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	return ctx, w
}

func TestOk(t *testing.T) {
	ctx, w := testContext()
	Ok(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Ok: status = %d, want %d", w.Code, http.StatusOK)
	}
	var body CommonResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Ok: decode: %v", err)
	}
	if body.Code != 2000001 {
		t.Errorf("Ok: code = %d, want 2000001", body.Code)
	}
	if body.Message != "success" {
		t.Errorf("Ok: message = %q, want %q", body.Message, "success")
	}
}

func TestOkWithMessage(t *testing.T) {
	ctx, w := testContext()
	OkWithMessage(ctx, "custom")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "custom" {
		t.Errorf("message = %q, want %q", body.Message, "custom")
	}
}

func TestOkWithData(t *testing.T) {
	ctx, w := testContext()
	OkWithData(ctx, map[string]int{"a": 1})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Data == nil {
		t.Error("data should not be nil")
	}
}

func TestResultWithCode(t *testing.T) {
	ctx, w := testContext()
	ResultWithCode(ctx, http.StatusCreated, 2010302, "data", "created")
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != 2010302 || body.Message != "created" || body.Data != "data" {
		t.Errorf("body = %+v", body)
	}
}

func TestCreated(t *testing.T) {
	ctx, w := testContext()
	Created(ctx, ServiceCodeUser, map[string]string{"id": "1"}, "")
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Resource created successfully" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestCreated_WithMessage(t *testing.T) {
	ctx, w := testContext()
	Created(ctx, "04", nil, "User created")
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "User created" {
		t.Errorf("message = %q, want %q", body.Message, "User created")
	}
}

func TestUpdated(t *testing.T) {
	ctx, w := testContext()
	Updated(ctx, ServiceCodeUser, nil, "")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Resource updated successfully" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestDeleted(t *testing.T) {
	ctx, w := testContext()
	Deleted(ctx, ServiceCodeUser, "")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Resource deleted successfully" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestFail(t *testing.T) {
	ctx, w := testContext()
	Fail(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "failure" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestFailWithMessage(t *testing.T) {
	ctx, w := testContext()
	FailWithMessage(ctx, "db error")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "db error" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestValidationError(t *testing.T) {
	ctx, w := testContext()
	ValidationError(ctx, ServiceCodeCommon, errors.New("validation failed"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
	var body ValidationErrorResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "The given data was invalid." {
		t.Errorf("message = %q", body.Message)
	}
	if body.Errors == nil {
		t.Error("errors should not be nil")
	}
}

func TestValidationErrorWithMessage(t *testing.T) {
	ctx, w := testContext()
	ValidationErrorWithMessage(ctx, "00", "Invalid input", map[string][]string{"x": {"err"}})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d", w.Code)
	}
	var body ValidationErrorResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Invalid input" {
		t.Errorf("message = %q", body.Message)
	}
	if len(body.Errors["x"]) != 1 || body.Errors["x"][0] != "err" {
		t.Errorf("errors = %v", body.Errors)
	}
}

func TestValidationErrorWithMessage_NilErrors(t *testing.T) {
	ctx, w := testContext()
	ValidationErrorWithMessage(ctx, "00", "msg", nil)
	var body ValidationErrorResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Errors == nil {
		t.Error("errors should be non-nil (initialized to empty map)")
	}
}

func TestValidationErrorSimple(t *testing.T) {
	ctx, w := testContext()
	ValidationErrorSimple(ctx, "00", "email", "The email must be valid.")
	var body ValidationErrorResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Errors["email"][0] != "The email must be valid." {
		t.Errorf("errors = %v", body.Errors)
	}
}

func TestUnauthorizedError(t *testing.T) {
	ctx, w := testContext()
	UnauthorizedError(ctx, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Unauthorized" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestNotFoundError(t *testing.T) {
	ctx, w := testContext()
	NotFoundError(ctx, "04", "", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Resource not found" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestConflictError(t *testing.T) {
	ctx, w := testContext()
	ConflictError(ctx, "04", "Email exists")
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "Email exists" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestForbiddenError(t *testing.T) {
	ctx, w := testContext()
	ForbiddenError(ctx, "No access")
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d", w.Code)
	}
	var body CommonResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Message != "No access" {
		t.Errorf("message = %q", body.Message)
	}
}

func TestCursorPaginated(t *testing.T) {
	next := "cursor123"
	ctx, w := testContext()
	CursorPaginated(ctx, http.StatusOK, ServiceCodeCommon, CaseCodeListRetrieved, CursorPaginationResponse{
		Data:       []int{1, 2},
		NextCursor: &next,
		HasNext:    true,
	}, "list")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
	var body CursorPaginatedResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.NextCursor == nil || *body.NextCursor != "cursor123" || !body.HasNext {
		t.Errorf("body = %+v", body)
	}
}

func TestSimplePaginated(t *testing.T) {
	ctx, w := testContext()
	SimplePaginated(ctx, http.StatusOK, ServiceCodeCommon, CaseCodeListRetrieved, SimplePaginationResponse{
		Data:       []string{"a"},
		PageNumber: 1,
		PageSize:   10,
		HasNext:    true,
		HasPrev:    false,
	}, "list")
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
	var body SimplePaginatedResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.PageNumber != 1 || body.PageSize != 10 || !body.HasNext || body.HasPrev {
		t.Errorf("body = %+v", body)
	}
}
