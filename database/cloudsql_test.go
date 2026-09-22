package database

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/turahe/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func testServiceAccountJSON(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	payload := map[string]string{
		"type":                        "service_account",
		"project_id":                  "test-project",
		"private_key_id":              "test-key",
		"private_key":                 string(pemKey),
		"client_email":                "test@test-project.iam.gserviceaccount.com",
		"client_id":                   "123456789",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com",
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return b
}

func installTestCloudSQLAuth(t *testing.T) {
	t.Helper()
	creds := testServiceAccountJSON(t)
	origDialer := newCloudSQLDialer
	origReg := registerCloudSQLMySQL
	newCloudSQLDialer = func(ctx context.Context, opts ...cloudsqlconn.Option) (*cloudsqlconn.Dialer, error) {
		opts = append([]cloudsqlconn.Option{cloudsqlconn.WithCredentialsJSON(creds)}, opts...)
		return cloudsqlconn.NewDialer(ctx, opts...)
	}
	registerCloudSQLMySQL = func(name string, opts ...cloudsqlconn.Option) (func() error, error) {
		opts = append([]cloudsqlconn.Option{cloudsqlconn.WithCredentialsJSON(creds)}, opts...)
		return origReg(name, opts...)
	}
	t.Cleanup(func() {
		newCloudSQLDialer = origDialer
		registerCloudSQLMySQL = origReg
	})
}

func stubCloudSQLGORMWithMock(t *testing.T) {
	t.Helper()
	origPG := openCloudSQLPostgresGORM
	origMy := openCloudSQLMySQLGORM
	openCloudSQLPostgresGORM = func(_ *sql.DB, opts *Options) (*gorm.DB, error) {
		sqlDB, mock, err := sqlmock.New()
		if err != nil {
			return nil, err
		}
		mock.ExpectClose()
		t.Cleanup(func() { _ = sqlDB.Close() })
		return gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{
			Logger: newFintechLogger(opts),
		})
	}
	openCloudSQLMySQLGORM = func(_ *config.DatabaseConfiguration, opts *Options, _ string) (*gorm.DB, error) {
		sqlDB, mock, err := sqlmock.New()
		if err != nil {
			return nil, err
		}
		mock.ExpectClose()
		t.Cleanup(func() { _ = sqlDB.Close() })
		return gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{
			Logger: newFintechLogger(opts),
		})
	}
	t.Cleanup(func() {
		openCloudSQLPostgresGORM = origPG
		openCloudSQLMySQLGORM = origMy
	})
}

func TestBuildDialerOptions(t *testing.T) {
	if len(buildDialerOptions(&Options{})) != 0 {
		t.Fatal("want 0")
	}
	if len(buildDialerOptions(&Options{UseIAM: true})) != 1 {
		t.Fatal("want 1 iam")
	}
	if len(buildDialerOptions(&Options{UsePrivateIP: true})) != 1 {
		t.Fatal("want 1 private")
	}
	if len(buildDialerOptions(&Options{UseIAM: true, UsePrivateIP: true})) != 2 {
		t.Fatal("want 2")
	}
}

func TestConnectCloudSQLPostgres_MissingInstance(t *testing.T) {
	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectCloudSQLMySQL_MissingInstance(t *testing.T) {
	_, _, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectCloudSQLPostgres_InvalidTimezone(t *testing.T) {
	installTestCloudSQLAuth(t)
	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
		ConnectionTimezone: "Not/AZone",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected timezone error")
	}
}

func TestConnectCloudSQLMySQL_InvalidTimezone(t *testing.T) {
	_, _, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
		ConnectionTimezone: "Not/AZone",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected timezone error")
	}
}

func TestConnectCloudSQLPostgres_DialerError(t *testing.T) {
	orig := newCloudSQLDialer
	newCloudSQLDialer = func(context.Context, ...cloudsqlconn.Option) (*cloudsqlconn.Dialer, error) {
		return nil, errors.New("dialer boom")
	}
	defer func() { newCloudSQLDialer = orig }()
	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected dialer error")
	}
}

func TestConnectCloudSQLMySQL_0_RegisterError(t *testing.T) {
	mysqlDriverMu.Lock()
	wasReady := mysqlDriverReady
	mysqlDriverReady = false
	mysqlDriverMu.Unlock()
	defer func() {
		mysqlDriverMu.Lock()
		mysqlDriverReady = wasReady
		mysqlDriverMu.Unlock()
	}()

	orig := registerCloudSQLMySQL
	registerCloudSQLMySQL = func(string, ...cloudsqlconn.Option) (func() error, error) {
		return nil, errors.New("register boom")
	}
	defer func() { registerCloudSQLMySQL = orig }()

	_, _, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected register error")
	}
}

func TestConnectCloudSQLPostgres_SuccessWithStubbedPing(t *testing.T) {
	installTestCloudSQLAuth(t)
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	db, cleanup, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
	}, &Options{
		UseIAM: true, UsePrivateIP: true, PingTimeout: time.Second,
		MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLife: time.Minute, ConnMaxIdle: time.Minute,
	})
	if err != nil {
		t.Fatalf("connectCloudSQLPostgres: %v", err)
	}
	if db == nil || cleanup == nil {
		t.Fatal("expected db and cleanup")
	}
	if err := cleanup(); err != nil {
		t.Errorf("cleanup: %v", err)
	}
}

func TestConnectCloudSQLPostgres_PingError(t *testing.T) {
	installTestCloudSQLAuth(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return errors.New("ping boom") }
	defer func() { pingSQLDB = origPing }()

	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected ping error")
	}
}

func TestConnectCloudSQLMySQL_SuccessWithStubbedPing(t *testing.T) {
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverCleanup = func() error { return nil }
	mysqlDriverMu.Unlock()

	db, cleanup, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{
		UseIAM: true, UsePrivateIP: true, PingTimeout: time.Second,
		MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLife: time.Minute, ConnMaxIdle: time.Minute,
	})
	if err != nil {
		t.Fatalf("connectCloudSQLMySQL: %v", err)
	}
	if db == nil || cleanup == nil {
		t.Fatal("expected db and cleanup")
	}
	if err := cleanup(); err != nil {
		t.Errorf("cleanup: %v", err)
	}
}

func TestConnectCloudSQLMySQL_PingError(t *testing.T) {
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return errors.New("ping boom") }
	defer func() { pingSQLDB = origPing }()

	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverMu.Unlock()

	_, _, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected ping error")
	}
}

func TestConnectCloudSQLPostgres_PasswordAuthPublicIP(t *testing.T) {
	installTestCloudSQLAuth(t)
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	db, cleanup, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
	}, &Options{PingTimeout: time.Second})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	_ = cleanup()
	if db == nil {
		t.Fatal("nil db")
	}
}

func TestConnectCloudSQLMySQL_PasswordAuthPublicIP(t *testing.T) {
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverCleanup = func() error { return nil }
	mysqlDriverMu.Unlock()

	db, cleanup, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	_ = cleanup()
	if db == nil {
		t.Fatal("nil db")
	}
}

func TestConnectCloudSQLPostgres_GORMOpenError(t *testing.T) {
	installTestCloudSQLAuth(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()
	orig := openCloudSQLPostgresGORM
	openCloudSQLPostgresGORM = func(*sql.DB, *Options) (*gorm.DB, error) {
		return nil, errors.New("gorm boom")
	}
	defer func() { openCloudSQLPostgresGORM = orig }()

	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected gorm open error")
	}
}

func TestConnectCloudSQLMySQL_CloseSQLError(t *testing.T) {
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverCleanup = nil
	mysqlDriverMu.Unlock()

	_, cleanup, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	// Close twice: second close of sql.DB returns error on some drivers; at least exercise closeFn.
	_ = cleanup()
	_ = cleanup()
}
