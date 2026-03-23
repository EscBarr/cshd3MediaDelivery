package storage

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, key string, path *string, file io.Reader) error
	Get(ctx context.Context, key string, path *string) (io.ReadSeeker, error)
	Delete(ctx context.Context, key string, path *string) error
}
