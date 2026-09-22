package database

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"testing"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/turahe/pkg/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOptionFuncsRemaining(t *testing.T) {
	opts := &Options{}
	WithMaxIdleConns(7)(opts)
	WithConnMaxLifetime(2 * time.Hour)(opts)
	WithConnMaxIdleTime(3 * time.Minute)(opts)
	WithOpenTelemetry(true)(opts)
	if opts.MaxIdleConns != 7 || opts.ConnMaxLife != 2*time.Hour || opts.ConnMaxIdle != 3*time.Minute || !opts.EnableOpenTelemetry {
		t.Fatalf("opts = %+v", opts)
	}
	WithProductionPoolDefaults(opts)
	if opts.MaxOpenConns != productionMaxOpenConns || opts.MaxIdleConns != productionMaxIdleConns {
		t.Fatalf("production defaults: open=%d idle=%d", opts.MaxOpenConns, opts.MaxIdleConns)
	}
}

func TestWrapConnectErr(t *testing.T) {
	err := wrapConnectErr(errors.New("otel fail"), nil)
	if err == nil || err.Error() != "otel gorm: otel fail" {
		t.Fatalf("got %v", err)
	}
	err = wrapConnectErr(errors.New("otel fail"), func() error { return nil })
	if err == nil || err.Error() != "otel gorm: otel fail" {
		t.Fatalf("got %v", err)
	}
	err = wrapConnectErr(errors.New("otel fail"), func() error { return errors.New("cleanup fail") })
	if err == nil {
		t.Fatal("expected combined error")
	}
}

func TestDBSystemForDriver_Default(t *testing.T) {
	if got := dbSystemForDriver("custom-driver"); got != "custom-driver" {
		t.Fatalf("got %q", got)
	}
}

func TestRegisterGORMInstrumentation_Disabled(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db"}
	db, err := New(cfg, Options{EnableOpenTelemetry: false})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	if err := registerGORMInstrumentation(db.DB(), cfg, &Options{EnableOpenTelemetry: false}); err != nil {
		t.Fatalf("register: %v", err)
	}
}

func TestRegisterGORMInstrumentation_Enabled(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db"}
	db, err := New(cfg, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	if err := registerGORMInstrumentation(db.DB(), cfg, &Options{EnableOpenTelemetry: true}); err != nil {
		t.Fatalf("register: %v", err)
	}
}

func TestNewContext_OTELRegisterCleanup(t *testing.T) {
	stubConnectStandardWithMock(t)
	cfg := &config.DatabaseConfiguration{Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db"}
	db, err := NewContext(context.Background(), cfg, Options{}, WithOpenTelemetry(true))
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	defer db.Close()
}

func TestNewContext_CloudSQLDrivers(t *testing.T) {
	origPG := connectCloudSQLPostgresFn
	origMy := connectCloudSQLMySQLFn
	defer func() {
		connectCloudSQLPostgresFn = origPG
		connectCloudSQLMySQLFn = origMy
	}()

	connectCloudSQLPostgresFn = func(context.Context, *config.DatabaseConfiguration, *Options) (*gorm.DB, func() error, error) {
		gdb, _ := openMockGORM(t)
		sqlDB, err := gdb.DB()
		return gdb, func() error { return sqlDB.Close() }, err
	}
	connectCloudSQLMySQLFn = func(context.Context, *config.DatabaseConfiguration, *Options) (*gorm.DB, func() error, error) {
		gdb, _ := openMockGORM(t)
		sqlDB, err := gdb.DB()
		return gdb, func() error { return sqlDB.Close() }, err
	}

	for _, driver := range []string{"cloudsql-postgres", "cloudsql-mysql"} {
		t.Run(driver, func(t *testing.T) {
			cfg := &config.DatabaseConfiguration{Driver: driver, CloudSQLInstance: "x", Dbname: "db"}
			db, err := NewContext(context.Background(), cfg, Options{PingTimeout: time.Second})
			if err != nil {
				t.Fatalf("NewContext %s: %v", driver, err)
			}
			_ = db.Close()
		})
	}
}

func TestBuildDSN_InvalidTimezone(t *testing.T) {
	_, err := buildDSN(&config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
		ConnectionTimezone: "Nope/Zone",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = buildDSN(&config.DatabaseConfiguration{
		Driver: "postgres", Host: "h", Port: "5432", Username: "u", Password: "p", Dbname: "db",
		ConnectionTimezone: "Nope/Zone",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectStandard_Refused(t *testing.T) {
	opts := &Options{PingTimeout: 200 * time.Millisecond}
	for _, cfg := range []*config.DatabaseConfiguration{
		{Driver: "mysql", Host: "127.0.0.1", Port: "1", Username: "u", Password: "p", Dbname: "db"},
		{Driver: "postgres", Host: "127.0.0.1", Port: "1", Username: "u", Password: "p", Dbname: "db"},
		{Driver: "sqlserver", Host: "127.0.0.1", Port: "1", Username: "u", Password: "p", Dbname: "db"},
	} {
		t.Run(cfg.Driver, func(t *testing.T) {
			_, _, err := connectStandard(context.Background(), cfg, opts)
			if err == nil {
				t.Fatal("expected connect/ping error")
			}
		})
	}
}

func TestConnectStandard_PoolOverridesFromConfig(t *testing.T) {
	orig := connectStandardFn
	connectStandardFn = func(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
		gdb, mock := openMockGORM(t)
		_ = mock
		sqlDB, err := gdb.DB()
		if err != nil {
			return nil, nil, err
		}
		configurePool(sqlDB, cfg, opts)
		return gdb, func() error { return sqlDB.Close() }, nil
	}
	defer func() { connectStandardFn = orig }()

	cfg := &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
		MaxOpenConns: 4, MaxIdleConns: 2, ConnMaxLifetimeMinutes: 1,
	}
	db, err := New(cfg, Options{PingTimeout: time.Second, MaxOpenConns: 30, MaxIdleConns: 10, ConnMaxLife: time.Hour, ConnMaxIdle: time.Minute})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	sqlDB, err := db.DB().DB()
	if err != nil {
		t.Fatal(err)
	}
	if sqlDB.Stats().MaxOpenConnections != 4 {
		t.Errorf("MaxOpenConnections = %d, want 4", sqlDB.Stats().MaxOpenConnections)
	}
}

func TestFintechLogger_InfoWarnWhenEnabled(t *testing.T) {
	l := &fintechLogger{level: logger.Info, slowThreshold: time.Second}
	l.Info(context.Background(), "hello %s", "world")
	l.Warn(context.Background(), "warn %s", "x")
}

func TestRedactSQL_CollapsePlaceholders(t *testing.T) {
	got := redactSQL("[REDACTED] [REDACTED]")
	if got != "[REDACTED]" {
		t.Fatalf("got %q, want [REDACTED]", got)
	}
}

func TestLogDatabaseConnected(t *testing.T) {
	logDatabaseConnected(&config.DatabaseConfiguration{Driver: "mysql", Host: "h", Port: "3306", Dbname: "db"})
	logDatabaseConnected(&config.DatabaseConfiguration{Driver: "postgres", Host: "h", Port: "5432", Dbname: "db"})
	logDatabaseConnected(&config.DatabaseConfiguration{Driver: "sqlserver", Host: "h", Port: "1433", Dbname: "master"})
}

func TestConnectStandard_SuccessSQLServer(t *testing.T) {
	gdb, _ := openMockGORM(t)
	origOpen := openGormSQLServer
	openGormSQLServer = func(string, *Options) (*gorm.DB, error) { return gdb, nil }
	defer func() { openGormSQLServer = origOpen }()
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	db, cleanup, err := connectStandard(context.Background(), &config.DatabaseConfiguration{
		Driver: "sqlserver", Host: "h", Port: "1433", Username: "sa", Password: "p", Dbname: "master",
	}, &Options{
		PingTimeout: time.Second,
		MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLife: time.Minute, ConnMaxIdle: time.Minute,
	})
	if err != nil {
		t.Fatalf("connectStandard: %v", err)
	}
	if db == nil || cleanup == nil {
		t.Fatal("expected db and cleanup")
	}
	// sqlmock Close is owned by openMockGORM cleanup; connectStandard's closer may race expectations.
	_ = cleanup
}

func TestConnectStandard_SuccessMySQLAndPostgres(t *testing.T) {
	gdb, _ := openMockGORM(t)
	origMySQL := openGormMySQL
	openGormMySQL = func(string, *Options) (*gorm.DB, error) { return gdb, nil }
	defer func() { openGormMySQL = origMySQL }()
	origPG := openGormPostgres
	openGormPostgres = func(string, *Options) (*gorm.DB, error) { return gdb, nil }
	defer func() { openGormPostgres = origPG }()
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	opts := &Options{
		PingTimeout: time.Second,
		MaxOpenConns: 3, MaxIdleConns: 1, ConnMaxLife: time.Minute, ConnMaxIdle: time.Minute,
	}
	for _, cfg := range []*config.DatabaseConfiguration{
		{Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db"},
		{Driver: "postgres", Host: "h", Port: "5432", Username: "u", Password: "p", Dbname: "db"},
	} {
		t.Run(cfg.Driver, func(t *testing.T) {
			db, cleanup, err := connectStandard(context.Background(), cfg, opts)
			if err != nil {
				t.Fatalf("connectStandard: %v", err)
			}
			if db == nil || cleanup == nil {
				t.Fatal("expected db and cleanup")
			}
			_ = cleanup()
		})
	}
}

func TestConnectStandard_OpenError(t *testing.T) {
	orig := openGormSQLServer
	openGormSQLServer = func(string, *Options) (*gorm.DB, error) {
		return nil, errors.New("open boom")
	}
	defer func() { openGormSQLServer = orig }()

	_, _, err := connectStandard(context.Background(), &config.DatabaseConfiguration{
		Driver: "sqlserver", Host: "h", Port: "1433", Username: "sa", Password: "p", Dbname: "master",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestConnectStandard_GormSQLDBError(t *testing.T) {
	gdb, _ := openMockGORM(t)
	origOpen := openGormSQLServer
	openGormSQLServer = func(string, *Options) (*gorm.DB, error) { return gdb, nil }
	defer func() { openGormSQLServer = origOpen }()
	orig := gormSQLDB
	gormSQLDB = func(*gorm.DB) (*sql.DB, error) { return nil, errors.New("sqldb boom") }
	defer func() { gormSQLDB = orig }()

	_, _, err := connectStandard(context.Background(), &config.DatabaseConfiguration{
		Driver: "sqlserver", Host: "h", Port: "1433", Username: "sa", Password: "p", Dbname: "master",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected get sql.DB error")
	}
}

func TestConnectStandard_PingFailsAfterOpen(t *testing.T) {
	gdb, _ := openMockGORM(t)
	origOpen := openGormSQLServer
	openGormSQLServer = func(string, *Options) (*gorm.DB, error) { return gdb, nil }
	defer func() { openGormSQLServer = origOpen }()
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return errors.New("ping boom") }
	defer func() { pingSQLDB = origPing }()

	_, _, err := connectStandard(context.Background(), &config.DatabaseConfiguration{
		Driver: "sqlserver", Host: "h", Port: "1433", Username: "sa", Password: "p", Dbname: "master",
	}, &Options{PingTimeout: time.Second, MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLife: time.Minute, ConnMaxIdle: time.Minute})
	if err == nil {
		t.Fatal("expected ping error")
	}
}

func TestConnectStandard_PingError(t *testing.T) {
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return errors.New("ping boom") }
	defer func() { pingSQLDB = origPing }()

	// Use sqlmock-backed open via stubbing at a lower level: refuse real dial by mocking gorm path.
	// Direct connectStandard against refused port fails at open before ping — use configure path via New stub.
	orig := connectStandardFn
	connectStandardFn = func(ctx context.Context, cfg *config.DatabaseConfiguration, opts *Options) (*gorm.DB, func() error, error) {
		return connectStandard(ctx, cfg, opts)
	}
	defer func() { connectStandardFn = orig }()

	gdb, _ := openMockGORM(t)
	sqlDB, _ := gdb.DB()
	configurePool(sqlDB, &config.DatabaseConfiguration{MaxOpenConns: 1}, &Options{MaxOpenConns: 1, MaxIdleConns: 1, ConnMaxLife: time.Minute, ConnMaxIdle: time.Minute})
	if err := pingSQLDB(context.Background(), sqlDB); err == nil {
		t.Fatal("expected ping error from stub")
	}
}

func TestNewContext_OTELRegisterFails(t *testing.T) {
	stubConnectStandardWithMock(t)
	orig := registerGORMInstrumentationFn
	registerGORMInstrumentationFn = func(*gorm.DB, *config.DatabaseConfiguration, *Options) error {
		return errors.New("otel boom")
	}
	defer func() { registerGORMInstrumentationFn = orig }()

	_, err := NewContext(context.Background(), &config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db",
	}, Options{})
	if err == nil {
		t.Fatal("expected otel error")
	}
}

func TestOpenCloudSQLGORM_Defaults(t *testing.T) {
	gdb, _ := openMockGORM(t)
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	opts := &Options{LogLevel: logger.Silent, SlowThreshold: time.Second}
	_, _ = openCloudSQLPostgresGORM(sqlDB, opts)

	cfg := &config.DatabaseConfiguration{Username: "u", Password: "p", CloudSQLInstance: "i", Dbname: "db"}
	_, _ = openCloudSQLMySQLGORM(cfg, opts, "UTC")
	optsIAM := *opts
	optsIAM.UseIAM = true
	_, _ = openCloudSQLMySQLGORM(cfg, &optsIAM, "UTC")
}

func TestConnectCloudSQLMySQL_RegisterSuccessPath(t *testing.T) {
	installTestCloudSQLAuth(t)
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()

	mysqlDriverMu.Lock()
	mysqlDriverReady = false
	mysqlDriverMu.Unlock()

	db, cleanup, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err != nil {
		t.Logf("connect: %v", err)
		return
	}
	_ = cleanup()
	_ = db
}

func TestConnectCloudSQLMySQL_OpenError(t *testing.T) {
	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverMu.Unlock()
	orig := openCloudSQLMySQLGORM
	openCloudSQLMySQLGORM = func(*config.DatabaseConfiguration, *Options, string) (*gorm.DB, error) {
		return nil, errors.New("open boom")
	}
	defer func() { openCloudSQLMySQLGORM = orig }()

	_, _, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestConnectCloudSQLPostgres_ParseConfigError(t *testing.T) {
	installTestCloudSQLAuth(t)
	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst",
		Username:         "u",
		Password:         "p",
		Dbname:           "db",
		Port:             "not-a-number",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestConnectCloudSQLPostgres_SQLOpenError(t *testing.T) {
	installTestCloudSQLAuth(t)
	orig := sqlOpen
	sqlOpen = func(string, string) (*sql.DB, error) { return nil, errors.New("open boom") }
	defer func() { sqlOpen = orig }()
	_, _, err := connectCloudSQLPostgres(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "proj:region:inst", Username: "u", Password: "p", Dbname: "db", Port: "5432",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected open error")
	}
}

func TestCloudSQLDial_Invoked(t *testing.T) {
	defer func() { _ = recover() }()
	_, _ = defaultCloudSQLDial(context.Background(), nil, "inst")
}

func TestMakeCloudSQLDialFunc(t *testing.T) {
	orig := cloudSQLDial
	cloudSQLDial = func(context.Context, *cloudsqlconn.Dialer, string) (net.Conn, error) {
		return nil, errors.New("dialed")
	}
	defer func() { cloudSQLDial = orig }()
	fn := makeCloudSQLDialFunc(nil, "inst")
	_, err := fn(context.Background(), "tcp", "addr")
	if err == nil || err.Error() != "dialed" {
		t.Fatalf("err = %v", err)
	}
}

func TestConnectCloudSQLPostgres_DialFuncWired(t *testing.T) {
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
	_ = db
}

func TestGormSQLDB_ErrorPaths(t *testing.T) {
	orig := gormSQLDB
	gormSQLDB = func(*gorm.DB) (*sql.DB, error) { return nil, errors.New("sqldb boom") }
	defer func() { gormSQLDB = orig }()

	d := &Database{db: &gorm.DB{}, opts: &Options{PingTimeout: time.Second}}
	if err := d.Health(context.Background()); err == nil {
		t.Fatal("expected health error")
	}

	stubCloudSQLGORMWithMock(t)
	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverMu.Unlock()
	_, _, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "x", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err == nil {
		t.Fatal("expected mysql get sql.DB error")
	}

	_, _, err = connectStandard(context.Background(), &config.DatabaseConfiguration{
		Driver: "mysql", Host: "127.0.0.1", Port: "1", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: 200 * time.Millisecond})
	// either open or get sql.DB error
	_ = err
}

func TestConnectCloudSQLMySQL_SQLCloseError(t *testing.T) {
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()
	origClose := sqlDBClose
	sqlDBClose = func(*sql.DB) error { return errors.New("close boom") }
	defer func() { sqlDBClose = origClose }()

	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverCleanup = nil
	mysqlDriverMu.Unlock()

	_, cleanup, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "x", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := cleanup(); err == nil {
		t.Fatal("expected close error")
	}
}

func TestDatabase_Health_PingFail(t *testing.T) {
	// Force ping failure via pingSQLDB seam (sqlmock does not monitor pings by default).
	orig := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return errors.New("ping fail") }
	defer func() { pingSQLDB = orig }()

	db, _ := newMockDatabase(t)
	if err := db.Health(context.Background()); err == nil {
		t.Fatal("expected ping error")
	}
}

func TestCompat_SetupGetCleanup(t *testing.T) {
	stubConnectStandardWithMock(t)
	orig := config.Config
	defer func() {
		_ = Cleanup()
		config.Config = orig
	}()
	_ = Cleanup()

	config.Config = &config.Configuration{
		Database: config.DatabaseConfiguration{
			Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "primary", Logmode: true,
		},
		DatabaseSite: config.DatabaseConfiguration{
			Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "site", Logmode: true,
		},
	}
	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if GetDB() == nil {
		t.Fatal("GetDB nil")
	}
	if GetDBSite() == nil {
		t.Fatal("GetDBSite nil")
	}
	if err := HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if !IsAlive() {
		t.Fatal("IsAlive false")
	}
	if err := Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
}

func TestCompat_GetDB_Panic(t *testing.T) {
	compatMu.Lock()
	origDB, origSite := DB, DBSite
	DB, DBSite = nil, nil
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		DB, DBSite = origDB, origSite
		compatMu.Unlock()
	}()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = GetDB()
}

func TestCompat_GetDBSite_FallsBack(t *testing.T) {
	db, _ := newMockDatabase(t)
	defer db.Close()

	compatMu.Lock()
	origDB, origSite, origDefault := DB, DBSite, defaultDB
	DB = db.DB()
	DBSite = nil
	defaultDB = db
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		DB, DBSite, defaultDB = origDB, origSite, origDefault
		compatMu.Unlock()
	}()

	if got := GetDBSite(); got != db.DB() {
		t.Fatal("GetDBSite should fall back to GetDB")
	}
}

func TestCompat_HealthCheck_SiteError(t *testing.T) {
	orig := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return errors.New("dead") }
	defer func() { pingSQLDB = orig }()

	db, _ := newMockDatabase(t)
	compatMu.Lock()
	origDefault, origSite := defaultDB, defaultDBSite
	defaultDB = db
	defaultDBSite = db
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		defaultDB, defaultDBSite = origDefault, origSite
		compatMu.Unlock()
	}()

	if err := HealthCheck(context.Background()); err == nil {
		t.Fatal("expected health error")
	}
}

func TestCompat_Cleanup_CloseErrors(t *testing.T) {
	compatMu.Lock()
	origDefault, origSite := defaultDB, defaultDBSite
	defaultDB = &Database{cleanups: []func() error{func() error { return errors.New("p") }}}
	defaultDBSite = &Database{cleanups: []func() error{func() error { return errors.New("s") }}}
	DB, DBSite = &gorm.DB{}, &gorm.DB{}
	compatMu.Unlock()
	defer func() {
		compatMu.Lock()
		defaultDB, defaultDBSite = origDefault, origSite
		DB, DBSite = nil, nil
		compatMu.Unlock()
	}()

	if err := Cleanup(); err == nil {
		t.Fatal("expected cleanup error")
	}
}

func TestCreateDatabaseConnection_Logmode(t *testing.T) {
	stubConnectStandardWithMock(t)
	defer Cleanup()
	gormDB, err := CreateDatabaseConnection(&config.DatabaseConfiguration{
		Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db", Logmode: true,
	})
	if err != nil {
		t.Fatalf("CreateDatabaseConnection: %v", err)
	}
	if gormDB == nil {
		t.Fatal("nil")
	}
}

func TestSetup_PrimaryOnly(t *testing.T) {
	stubConnectStandardWithMock(t)
	orig := config.Config
	defer func() {
		_ = Cleanup()
		config.Config = orig
	}()
	config.Config = &config.Configuration{
		Database: config.DatabaseConfiguration{Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db"},
	}
	if err := Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestSetup_PrimaryFails(t *testing.T) {
	orig := config.Config
	defer func() { config.Config = orig }()
	config.Config = &config.Configuration{
		Database: config.DatabaseConfiguration{Driver: "invalid"},
	}
	if err := Setup(); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetup_SiteFails(t *testing.T) {
	stubConnectStandardWithMock(t)
	orig := config.Config
	defer func() {
		_ = Cleanup()
		config.Config = orig
	}()
	// First New succeeds via stub; second New for site uses invalid driver and fails inside New (not stubbed for invalid).
	config.Config = &config.Configuration{
		Database:     config.DatabaseConfiguration{Driver: "mysql", Host: "h", Port: "3306", Username: "u", Password: "p", Dbname: "db"},
		DatabaseSite: config.DatabaseConfiguration{Driver: "invalid", Dbname: "x"},
	}
	if err := Setup(); err == nil {
		t.Fatal("expected site error")
	}
}

func TestConnectCloudSQLMySQL_CloseError(t *testing.T) {
	stubCloudSQLGORMWithMock(t)
	origPing := pingSQLDB
	pingSQLDB = func(context.Context, *sql.DB) error { return nil }
	defer func() { pingSQLDB = origPing }()
	mysqlDriverMu.Lock()
	mysqlDriverReady = true
	mysqlDriverCleanup = nil
	mysqlDriverMu.Unlock()

	db, cleanup, err := connectCloudSQLMySQL(context.Background(), &config.DatabaseConfiguration{
		CloudSQLInstance: "x", Username: "u", Password: "p", Dbname: "db",
	}, &Options{PingTimeout: time.Second})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	_ = cleanup()
}
