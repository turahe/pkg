package mtls

import (
	"crypto/tls"
	"strings"
	"testing"
)

func TestServerTLSConfig_disabled(t *testing.T) {
	_, err := ServerTLSConfig(Config{Enabled: false})
	if err == nil || !strings.Contains(err.Error(), "MTLS_ENABLED is false") {
		t.Fatalf("ServerTLSConfig() err = %v", err)
	}
}

func TestClientTLSConfig_disabled(t *testing.T) {
	_, err := ClientTLSConfig(Config{Enabled: false})
	if err == nil || !strings.Contains(err.Error(), "MTLS_ENABLED is false") {
		t.Fatalf("ClientTLSConfig() err = %v", err)
	}
}

func TestServerTLSConfig_success(t *testing.T) {
	cfg := testConfig(t)

	tlsCfg, err := ServerTLSConfig(cfg)
	if err != nil {
		t.Fatalf("ServerTLSConfig() err = %v", err)
	}
	if tlsCfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %x", tlsCfg.MinVersion)
	}
	if tlsCfg.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("ClientAuth = %v", tlsCfg.ClientAuth)
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Fatalf("Certificates len = %d", len(tlsCfg.Certificates))
	}
	if tlsCfg.ClientCAs == nil {
		t.Fatal("ClientCAs is nil")
	}
}

func TestClientTLSConfig_success(t *testing.T) {
	cfg := testConfig(t)

	tlsCfg, err := ClientTLSConfig(cfg)
	if err != nil {
		t.Fatalf("ClientTLSConfig() err = %v", err)
	}
	if tlsCfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %x", tlsCfg.MinVersion)
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Fatalf("Certificates len = %d", len(tlsCfg.Certificates))
	}
	if tlsCfg.RootCAs == nil {
		t.Fatal("RootCAs is nil")
	}
}

func TestServerTLSConfig_missingCA(t *testing.T) {
	cfg := testConfig(t)
	cfg.CACertFile = filepathMissing(t)

	_, err := ServerTLSConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "read CA cert") {
		t.Fatalf("ServerTLSConfig() err = %v", err)
	}
}

func TestServerTLSConfig_invalidCA(t *testing.T) {
	cfg := testConfig(t)
	writeFile(t, cfg.CACertFile, []byte("not a pem"))

	_, err := ServerTLSConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "invalid CA PEM") {
		t.Fatalf("ServerTLSConfig() err = %v", err)
	}
}

func filepathMissing(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/missing.crt"
}
