# Getting Started

## Requirements

- Go **1.26+**
- Optional: Docker / Docker Compose for Redis, MySQL, Postgres (tests and local services)

## Install

```bash
go get github.com/turahe/pkg
```

Copy [`.env.example`](../.env.example) to `.env` and fill required values (`DATABASE_*`, JWT key paths, etc.).

## Minimal wiring

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turahe/pkg/config"
	"github.com/turahe/pkg/database"
	"github.com/turahe/pkg/middlewares"
	"github.com/turahe/pkg/otelx"
	"github.com/turahe/pkg/sentryx"
	"gorm.io/gorm/logger"
)

func main() {
	if err := config.Setup(""); err != nil {
		log.Fatal(err)
	}
	cfg := config.GetConfig()

	otelShutdown, _ := otelx.Init(context.Background(), cfg.OpenTelemetry)
	defer func() { _ = otelShutdown(context.Background()) }()

	_ = sentryx.Init(cfg.Sentry)
	defer sentryx.Flush(cfg.Sentry.FlushTimeout)

	db, err := database.New(&cfg.Database, database.Options{LogLevel: logger.Warn})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.New()
	r.Use(
		middlewares.RecoveryHandler,
		middlewares.TraceMiddleware(),
		middlewares.LoggerMiddleware(),
		middlewares.Metrics(),
		middlewares.RequestTimeout(30*time.Second),
		middlewares.CORS(),
	)

	r.GET("/live", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/ready", func(c *gin.Context) {
		if err := db.Health(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
```

For the full production stack (Redis, mTLS, readiness gate, Prometheus `/metrics`), see the [Production Wiring Example](../README.md#production-wiring-example) in the root README.

## Recommended middleware order

```
Recovery → Trace → Logger → Metrics → Timeout → CORS → [mTLS] → [Auth] → [RateLimiter] → handlers
```

## Next steps

- [Architecture](architecture.md) — layer boundaries
- [Environment Variables](environment.md) — configure each subsystem
- [Packages](README.md#packages) — per-package guides
- Run tests: `make test` or `make test-docker` (see [Development](development.md))
