# Media Storage

Service for storing media files.

Supports two storage backends:

* filesystem
* S3 (MinIO)

## Run

```bash
docker compose up
```

## Storage configuration

Storage type is configured in:

```
config-yaml/docker.yaml
```

### Filesystem

To use filesystem storage leave `minio.endpoint` empty:

```yaml
MINIO_INFO:
  endpoint: ""
```

### S3 / MinIO

To use S3 storage set MinIO endpoint:

```yaml
MINIO_INFO:
  endpoint: "minio:9000"
```
