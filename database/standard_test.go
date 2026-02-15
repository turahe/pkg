package database

import (
	"strings"
	"testing"

	"github.com/turahe/pkg/config"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.DatabaseConfiguration
		wantErr bool
		contains []string
	}{
		{
			name: "mysql",
			cfg: &config.DatabaseConfiguration{
				Driver:   "mysql",
				Host:     "localhost",
				Port:     "3306",
				Username: "u",
				Password: "p",
				Dbname:   "db",
			},
			contains: []string{"tcp(localhost:3306)", "/db", "charset=utf8", "parseTime=True"},
		},
		{
			name: "postgres",
			cfg: &config.DatabaseConfiguration{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     "5432",
				Username: "u",
				Password: "p",
				Dbname:   "db",
				Sslmode:  false,
			},
			contains: []string{"host=localhost", "sslmode=disable", "dbname=db"},
		},
		{
			name: "postgres sslmode",
			cfg: &config.DatabaseConfiguration{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     "5432",
				Username: "u",
				Password: "p",
				Dbname:   "db",
				Sslmode:  true,
			},
			contains: []string{"sslmode=require"},
		},
		{
			name: "sqlite",
			cfg: &config.DatabaseConfiguration{
				Driver: "sqlite",
				Dbname: "mydb",
			},
			contains: []string{"./mydb.db"},
		},
		{
			name: "sqlserver",
			cfg: &config.DatabaseConfiguration{
				Driver:   "sqlserver",
				Host:     "localhost",
				Port:     "1433",
				Username: "u",
				Password: "p",
				Dbname:   "db",
				Sslmode:  false,
			},
			contains: []string{"sqlserver://", "encrypt=disable"},
		},
		{
			name: "sqlserver sslmode",
			cfg: &config.DatabaseConfiguration{
				Driver:   "sqlserver",
				Host:     "localhost",
				Port:     "1433",
				Username: "u",
				Password: "p",
				Dbname:   "db",
				Sslmode:  true,
			},
			contains: []string{"encrypt=true"},
		},
		{
			name: "unsupported driver",
			cfg: &config.DatabaseConfiguration{
				Driver: "invalid",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildDSN(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildDSN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			for _, sub := range tt.contains {
				if !strings.Contains(got, sub) {
					t.Errorf("buildDSN() = %q, want to contain %q", got, sub)
				}
			}
		})
	}
}
