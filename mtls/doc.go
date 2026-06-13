// Package mtls provides mutual TLS helpers for backend HTTPS and gateway outbound clients.
//
// Configuration is loaded via config.GetConfig().MTLS (or config.Setup). Set MTLS_ENABLED=true
// to require client certificates on the server and to attach a client cert on outbound HTTP.
package mtls
