package otelx

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/turahe/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTracingEnabled(t *testing.T) {
	if TracingEnabled(config.OpenTelemetryConfiguration{}) {
		t.Fatal("expected tracing disabled")
	}
	if !TracingEnabled(config.OpenTelemetryConfiguration{Endpoint: "localhost:4318"}) {
		t.Fatal("expected tracing enabled")
	}
}

func TestGORMEnabled(t *testing.T) {
	if GORMEnabled(config.OpenTelemetryConfiguration{GORMEnabled: true}) {
		t.Fatal("expected gorm disabled without endpoint")
	}
	cfg := config.OpenTelemetryConfiguration{
		GORMEnabled: true,
		Endpoint:    "localhost:4318",
	}
	if !GORMEnabled(cfg) {
		t.Fatal("expected gorm enabled")
	}
}

func TestRegisterGORM_MySQLMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()
	mock.ExpectPing()
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	if err := RegisterGORM(db, GORMOptions{DBSystem: "mysql"}); err != nil {
		t.Fatalf("RegisterGORM() error = %v", err)
	}
}
