package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func BenchmarkResolveActorType(b *testing.B) {
	inputs := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "admins", input: "admins"},
		{name: "User", input: "User"},
		{name: "system", input: "system"},
		{name: "unknown", input: "unknown-table"},
	}

	for _, tt := range inputs {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			var sink string
			for b.Loop() {
				sink = ResolveActorType(tt.input)
			}
			_ = sink
		})
	}
}

func BenchmarkManager_GenerateToken(b *testing.B) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	for _, alg := range []struct {
		name  string
		setup func(testing.TB) *Manager
	}{
		{name: "RS256", setup: setupRS256Config},
		{name: "ES256", setup: setupES256Config},
	} {
		b.Run(alg.name, func(b *testing.B) {
			m := alg.setup(b)
			b.ReportAllocs()
			var (
				token string
				err   error
			)
			for b.Loop() {
				token, err = m.GenerateToken(id, "users")
			}
			if err != nil {
				b.Fatal(err)
			}
			_ = token
		})
	}
}

func BenchmarkManager_ValidateToken(b *testing.B) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	for _, alg := range []struct {
		name  string
		setup func(testing.TB) *Manager
	}{
		{name: "RS256", setup: setupRS256Config},
		{name: "ES256", setup: setupES256Config},
	} {
		b.Run(alg.name, func(b *testing.B) {
			m := alg.setup(b)
			token, err := m.GenerateTokenWithExpiry(id, time.Hour, "admins")
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			var (
				claims *Claims
				verr   error
			)
			for b.Loop() {
				claims, verr = m.ValidateToken(token)
			}
			if verr != nil {
				b.Fatal(verr)
			}
			_ = claims
		})
	}
}

func BenchmarkGetActorType(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Set("actor_type", ActorTypeAdmin)

	b.ReportAllocs()
	var (
		got   string
		found bool
	)
	for b.Loop() {
		got, found = GetActorType(c)
	}
	if !found || got != ActorTypeAdmin {
		b.Fatalf("GetActorType = (%q, %v)", got, found)
	}
}
