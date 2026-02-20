# syntax=docker/dockerfile:1

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGo (required by sqlite and mssql drivers).
# Remove this if those drivers are not used in your service.
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

# Cache dependency downloads separately from source compilation.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a statically linked binary. -trimpath removes local file paths from
# stack traces (security), ldflags strip debug info (smaller binary).
RUN CGO_ENABLED=1 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd/example

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
# gcr.io/distroless/base-debian12 includes glibc (needed for CGo) but no shell,
# no package manager, and no root tools — minimal attack surface.
FROM gcr.io/distroless/base-debian12:nonroot

# Copy only the compiled binary.
COPY --from=builder /out/server /server

# Expose HTTP and metrics ports.
EXPOSE 8080

# nonroot user (UID 65532) is built into the distroless image.
USER nonroot:nonroot

# Kubernetes will override this with its own liveness/readiness probes,
# but HEALTHCHECK ensures local `docker run` surfaces health status.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "-health-check"] 

ENTRYPOINT ["/server"]
