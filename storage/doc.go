/*
Package storage provides a multi-driver object storage facade configured from the config package.

Role in architecture:
  - Infrastructure facade: exposes one process-wide storage driver and bucket for object read/write.

Responsibilities:
  - Setup / SetupContext: create a storage driver from config.Storage; empty STORAGE_DRIVER disables storage.
  - GetBucketName: access the configured default bucket name.
  - ReadObject, ReadObjectAsReader, WriteObject, DeleteObject, ObjectExists, ListObjects, PresignUpload: delegate to the configured driver; ctx is first parameter.
  - Close: close the configured driver.

Constraints:
  - Drivers are selected by STORAGE_DRIVER. Supported drivers are gcs, s3, and r2.
  - Do not store context in package state; pass ctx through every I/O call.
  - No business logic; only storage driver lifecycle and operation delegation.

This package must NOT:
  - Contain use-case logic; only storage operations.
*/
package storage
