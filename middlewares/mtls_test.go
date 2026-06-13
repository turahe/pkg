package middlewares

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/turahe/pkg/config"
)

func TestMTLSMiddleware_disabled(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		MTLS: config.MTLSConfiguration{Enabled: false},
	}

	router := setupRouter()
	router.Use(MTLSMiddleware())
	router.GET("/api", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMTLSMiddleware_missingCert(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		MTLS: config.MTLSConfiguration{
			Enabled:   true,
			SkipPaths: "/live",
		},
	}

	router := setupRouter()
	router.Use(MTLSMiddleware())
	router.GET("/api", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMTLSMiddleware_withCert(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		MTLS: config.MTLSConfiguration{Enabled: true},
	}

	router := setupRouter()
	router.Use(MTLSMiddleware())
	router.GET("/api", func(c *gin.Context) {
		cn, _ := c.Get(ContextMTLSClientCN)
		c.String(http.StatusOK, cn.(string))
	})

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{{Subject: pkix.Name{CommonName: "gateway"}}},
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gateway", w.Body.String())
}

func TestMTLSMiddleware_skipPath(t *testing.T) {
	original := config.Config
	defer func() { config.Config = original }()

	config.Config = &config.Configuration{
		MTLS: config.MTLSConfiguration{
			Enabled:   true,
			SkipPaths: "/live",
		},
	}

	router := setupRouter()
	router.Use(MTLSMiddleware())
	router.GET("/live", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
