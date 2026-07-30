# gcs

Google Cloud Storage client wrapper driven by config.

**Import:** `github.com/turahe/pkg/gcs`

## Lifecycle

```go
if err := gcs.Setup(); err != nil {
    log.Fatal(err)
}
defer gcs.Close()

client := gcs.GetClient()
bucket := gcs.GetBucket()
name := gcs.GetBucketName()
```

Credentials: `GCS_CREDENTIALS_FILE` or Application Default Credentials.

## Operations

```go
data, err := gcs.ReadObject(ctx, "path/object")
r, err := gcs.ReadObjectAsReader(ctx, "path/object")
err = gcs.WriteObject(ctx, "path/object", data, "application/octet-stream")
err = gcs.DeleteObject(ctx, "path/object")
exists, err := gcs.ObjectExists(ctx, "path/object")
names, err := gcs.ListObjects(ctx, "prefix/")
```

## Constraints

- Single client / bucket; no provider switching.
- No use-case logic.
- Disabled / no-op paths when `GCS_ENABLED` is false (see package behavior on Setup).

## See also

- [Environment — GCS](../environment.md#gcs)
