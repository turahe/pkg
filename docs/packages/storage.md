# storage

Multi-driver object storage facade driven by config.

**Import:** `github.com/turahe/pkg/storage`

## Lifecycle

```go
if err := storage.Setup(); err != nil { // or storage.SetupContext(ctx)
    log.Fatal(err)
}
defer storage.Close()

name := storage.GetBucketName()
```

Set `STORAGE_DRIVER` to `gcs`, `s3`, or `r2`. Leave it empty to disable storage; operations before setup or while disabled return `storage.ErrNotInitialized`.

## Drivers

| Driver | Required env | Notes |
|--------|--------------|-------|
| `gcs` | `STORAGE_BUCKET` | Uses Google Cloud Storage. `GCS_CREDENTIALS_FILE` is optional; omit it to use Application Default Credentials. |
| `s3` | `STORAGE_BUCKET`, `STORAGE_ENDPOINT`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY` | Generic S3-compatible driver. Defaults `STORAGE_REGION` to `us-east-1` and path-style requests to true. |
| `r2` | `STORAGE_BUCKET`, `STORAGE_ENDPOINT`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY` | Cloudflare R2 through the S3 API. Defaults `STORAGE_REGION` to `auto` and path-style requests to false. |

## Operations

```go
data, err := storage.ReadObject(ctx, "path/object")
r, err := storage.ReadObjectAsReader(ctx, "path/object")
err = storage.WriteObject(ctx, "path/object", data, "application/octet-stream")
err = storage.DeleteObject(ctx, "path/object")
exists, err := storage.ObjectExists(ctx, "path/object")
names, err := storage.ListObjects(ctx, "prefix/")
```

## Local RustFS

RustFS runs as the local S3-compatible service in Docker Compose:

```bash
docker compose up -d rustfs
```

Use these values from the host:

```bash
STORAGE_DRIVER=s3
STORAGE_BUCKET=pkg
STORAGE_ENDPOINT=http://127.0.0.1:9000
STORAGE_REGION=us-east-1
STORAGE_ACCESS_KEY=rustfsadmin
STORAGE_SECRET_KEY=rustfsadmin
STORAGE_FORCE_PATH_STYLE=true
```

Inside the compose network, use `STORAGE_ENDPOINT=http://rustfs:9000`.

Create the bucket once if it does not exist yet (for example `aws --endpoint-url http://127.0.0.1:9000 s3 mb s3://pkg`). Integration tests create `pkg` automatically.

## Cloudflare R2

```bash
STORAGE_DRIVER=r2
STORAGE_BUCKET=my-bucket
STORAGE_ENDPOINT=https://<accountid>.r2.cloudflarestorage.com
STORAGE_ACCESS_KEY=<r2-access-key>
STORAGE_SECRET_KEY=<r2-secret-key>
STORAGE_REGION=auto
STORAGE_FORCE_PATH_STYLE=false
```

## Migration from gcs

1. Replace `.../gcs` imports with `.../storage`.
2. Replace `gcs.Setup`, `gcs.Close`, and operation calls with `storage.Setup`, `storage.Close`, and matching `storage` operations.
3. Remove uses of `GetClient` and `GetBucket`; the new package exposes provider-neutral operations only.
4. Replace `GCS_ENABLED` and `GCS_BUCKET_NAME` with `STORAGE_DRIVER` and `STORAGE_BUCKET`.
5. For GCS, set `STORAGE_DRIVER=gcs`, `STORAGE_BUCKET`, and optionally `GCS_CREDENTIALS_FILE`.

## Constraints

- Single process-wide driver and bucket.
- No provider-specific public client handles.
- No use-case logic.

## See also

- [Environment — Storage](../environment.md#storage)
