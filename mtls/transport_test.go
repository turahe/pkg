package mtls

import (
	"crypto/tls"
	"testing"
)

func TestNewTransport_disabled(t *testing.T) {
	tr, err := NewTransport(Config{Enabled: false})
	if err != nil {
		t.Fatalf("NewTransport() err = %v", err)
	}
	if tr.TLSClientConfig != nil {
		t.Fatal("expected plain transport without TLSClientConfig")
	}
}

func TestNewTransport_enabled(t *testing.T) {
	cfg := testConfig(t)

	tr, err := NewTransport(cfg)
	if err != nil {
		t.Fatalf("NewTransport() err = %v", err)
	}
	if tr.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig on mTLS transport")
	}
	if tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %x", tr.TLSClientConfig.MinVersion)
	}
}

func TestCloneTLSConfig(t *testing.T) {
	if CloneTLSConfig(nil) != nil {
		t.Fatal("CloneTLSConfig(nil) should be nil")
	}

	orig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: "example"}
	cloned := CloneTLSConfig(orig)
	if cloned == orig {
		t.Fatal("CloneTLSConfig should return a copy")
	}
	if cloned.ServerName != orig.ServerName {
		t.Fatalf("ServerName = %q", cloned.ServerName)
	}
}
