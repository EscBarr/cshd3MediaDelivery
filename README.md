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
minio:
  endpoint: ""
```

### S3 / MinIO

To use S3 storage set MinIO endpoint:

```yaml
minio:
  endpoint: "minio:9000"
```
