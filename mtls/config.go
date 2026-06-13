package mtls

import "github.com/turahe/pkg/config"

// Config is the mTLS configuration loaded from the central config package.
type Config = config.MTLSConfiguration

// LoadConfig returns mTLS settings from the global config package.
func LoadConfig() Config {
	return config.GetConfig().MTLS
}
