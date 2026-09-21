/*
Package gcs provides a Google Cloud Storage client wrapper (cloud.google.com/go/storage) configured from config package.

Role in architecture:
  - Infrastructure adapter: connects to GCS, exposes client and bucket; used by application code for object read/write.

Responsibilities:
  - Setup / SetupContext: create storage client from config (credentials file or ADC); verify bucket access if BucketName set.
  - GetClient, GetBucket, GetBucketName: access the client and default bucket.
  - ReadObject, ReadObjectAsReader, WriteObject, DeleteObject, ObjectExists, ListObjects: map to GCS API calls; ctx is first parameter.
  - Close: close the client.

Constraints:
  - SDK: cloud.google.com/go/storage. Single client and bucket per process; no provider switching.
  - Do not store context in package state; pass ctx through every I/O call.
  - No business logic; only mapping to GCS operations.
  - Credentials: config.GCS.CredentialsFile or Application Default Credentials.

This package must NOT:
  - Contain use-case logic; only GCS operations.
*/
package gcs
