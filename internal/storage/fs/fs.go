package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type FSStorage struct {
	basePath string
}

func NewFSStorage(basePath string) *FSStorage {
	return &FSStorage{basePath: basePath}
}

func (f *FSStorage) Save(ctx context.Context, key string, file io.Reader) error {
	fullPath := filepath.Join(f.basePath, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

func (f *FSStorage) Get(ctx context.Context, key string) (io.ReadSeeker, error) {
	return os.Open(filepath.Join(f.basePath, key))
}

func (f *FSStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(filepath.Join(f.basePath, key))
}
