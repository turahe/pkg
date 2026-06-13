package mtls

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListenAndServe starts plain HTTP or mTLS HTTPS depending on cfg.Enabled.
func ListenAndServe(handler http.Handler, addr string) error {
	return ListenAndServeConfig(handler, addr, LoadConfig())
}

// ListenAndServeConfig starts plain HTTP or mTLS HTTPS using the given configuration.
func ListenAndServeConfig(handler http.Handler, addr string, cfg Config) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	if err := ConfigureServerWithConfig(srv, cfg); err != nil {
		return err
	}
	return ListenConfiguredWithConfig(srv, cfg)
}

// RunGin is a drop-in for gin.Engine.Run with optional mTLS.
func RunGin(engine *gin.Engine, addr string) error {
	return RunGinConfig(engine, addr, LoadConfig())
}

// RunGinConfig is RunGin with an explicit configuration.
func RunGinConfig(engine *gin.Engine, addr string, cfg Config) error {
	if err := ListenAndServeConfig(engine, addr, cfg); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// ConfigureServer applies mTLS to an existing http.Server when enabled in config.
func ConfigureServer(srv *http.Server) error {
	return ConfigureServerWithConfig(srv, LoadConfig())
}

// ConfigureServerWithConfig applies mTLS to srv when cfg.Enabled is true.
func ConfigureServerWithConfig(srv *http.Server, cfg Config) error {
	if !cfg.Enabled {
		return nil
	}
	tlsCfg, err := ServerTLSConfig(cfg)
	if err != nil {
		return fmt.Errorf("mtls configure server: %w", err)
	}
	srv.TLSConfig = tlsCfg
	return nil
}

// ListenConfigured serves using srv.TLSConfig when mTLS is enabled.
func ListenConfigured(srv *http.Server) error {
	return ListenConfiguredWithConfig(srv, LoadConfig())
}

// ListenConfiguredWithConfig serves with TLS when cfg.Enabled is true.
func ListenConfiguredWithConfig(srv *http.Server, cfg Config) error {
	if cfg.Enabled {
		return srv.ListenAndServeTLS(cfg.ServerCertFile, cfg.ServerKeyFile)
	}
	return srv.ListenAndServe()
}
