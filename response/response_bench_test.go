package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkBuildResponseCode(b *testing.B) {
	b.ReportAllocs()
	var sink int
	for b.Loop() {
		sink = BuildResponseCode(http.StatusOK, ServiceCodeCommon, CaseCodeSuccess)
	}
	_ = sink
}

func BenchmarkOkWithData(b *testing.B) {
	gin.SetMode(gin.TestMode)
	payload := map[string]any{"id": "550e8400-e29b-41d4-a716-446655440000", "name": "bench"}

	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		OkWithData(c, payload)
	}
}

func BenchmarkResult(b *testing.B) {
	gin.SetMode(gin.TestMode)
	payload := map[string]any{"ok": true}

	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		Result(c, http.StatusOK, ServiceCodeCommon, CaseCodeSuccess, payload, "success")
	}
}
