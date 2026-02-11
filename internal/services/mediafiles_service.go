package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"cshdMediaDelivery/internal/storage"
	"github.com/google/uuid"
)

type MediaService interface {
	Upload(ctx context.Context, file io.Reader, originalName string) (string, error)
	Get(ctx context.Context, key string) (io.ReadSeeker, error)
	Delete(ctx context.Context, key string) error
}

type mediaService struct {
	storage storage.Storage
}

func NewMediaService(storage storage.Storage) MediaService {
	return &mediaService{storage: storage}
}

func (m *mediaService) Upload(ctx context.Context, file io.Reader, originalName string) (string, error) {
	id := uuid.New().String()
	ext := filepath.Ext(originalName)

	key := fmt.Sprintf("%s%s", id, ext)

	if err := m.storage.Save(ctx, key, file); err != nil {
		return "", err
	}

	return key, nil
}

func (m *mediaService) Get(ctx context.Context, key string) (io.ReadSeeker, error) {
	return m.storage.Get(ctx, key)
}

func (m *mediaService) Delete(ctx context.Context, key string) error {
	return m.storage.Delete(ctx, key)
}
