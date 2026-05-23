package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/turahe/pkg/config"
)

// resolveConnectionTimezone returns the IANA timezone for the DB session.
// Uses cfg.ConnectionTimezone when set, otherwise config.Timezone.Timezone, then UTC.
func resolveConnectionTimezone(cfg *config.DatabaseConfiguration) (string, error) {
	tz := cfg.ConnectionTimezone
	if tz == "" {
		if c := config.GetConfig(); c != nil && c.Timezone.Timezone != "" {
			tz = c.Timezone.Timezone
		} else {
			tz = "UTC"
		}
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return "", fmt.Errorf("invalid connection timezone %q: %w", tz, err)
	}
	return tz, nil
}

func mysqlLocQueryValue(tz string) string {
	return url.QueryEscape(tz)
}
