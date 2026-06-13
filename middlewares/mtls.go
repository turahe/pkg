package middlewares

import (
	"crypto/x509"

	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	// ContextMTLSClientCN is the verified client certificate Common Name.
	ContextMTLSClientCN = "mtls_client_cn"
)

// MTLSMiddleware returns Gin middleware that requires a verified TLS client certificate
// when config.MTLS.Enabled is true. Skips paths listed in MTLS_SKIP_PATHS.
// Use with mtls.ConfigureServer so the listener verifies client certs at the TLS layer.
func MTLSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig().MTLS
	if !cfg.Enabled {
		return func(ctx *gin.Context) { ctx.Next() }
	}

	skipPaths := parseSkipPaths(cfg.SkipPaths)

	return func(ctx *gin.Context) {
		if shouldSkipPath(ctx.Request.URL.Path, skipPaths) {
			ctx.Next()
			return
		}

		cert, ok := peerCertificate(ctx)
		if !ok {
			response.ForbiddenError(ctx, "Client certificate required")
			ctx.Abort()
			return
		}

		ctx.Set(ContextMTLSClientCN, cert.Subject.CommonName)
		ctx.Next()
	}
}

func peerCertificate(ctx *gin.Context) (*x509.Certificate, bool) {
	tlsState := ctx.Request.TLS
	if tlsState == nil || len(tlsState.PeerCertificates) == 0 {
		return nil, false
	}
	return tlsState.PeerCertificates[0], true
}
