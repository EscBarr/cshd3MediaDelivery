package services

import (
	"bytes"
	"context"
	"cshdMediaDelivery/internal/lib/errs"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"cshdMediaDelivery/internal/storage"

	"github.com/google/uuid"
)

var allowedTypes = map[string]string{

	// Images
	"image/jpeg":    "jpg",
	"image/png":     "png",
	"image/webp":    "webp",
	"image/gif":     "gif",
	"image/bmp":     "bmp",
	"image/tiff":    "tiff",
	"image/svg+xml": "svg",

	// PDF
	"application/pdf": "pdf",

	// Word
	"application/msword": "doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",

	// Excel
	"application/vnd.ms-excel": "xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "xlsx",

	// PowerPoint
	"application/vnd.ms-powerpoint":                                             "ppt",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": "pptx",

	// OpenDocument
	"application/vnd.oasis.opendocument.text":         "odt",
	"application/vnd.oasis.opendocument.spreadsheet":  "ods",
	"application/vnd.oasis.opendocument.presentation": "odp",

	// CSV
	"text/csv": "csv",
}

func detectAllowedExtension(contentType, originalName string) (string, bool) {

	// прямое совпадение
	if ext, ok := allowedTypes[contentType]; ok {
		return ext, true
	}

	// Office файлы иногда определяются как zip
	if contentType == "application/zip" {
		ext := strings.ToLower(filepath.Ext(originalName))

		switch ext {
		case ".docx", ".xlsx", ".pptx":
			return strings.TrimPrefix(ext, "."), true
		}
	}

	return "", false
}

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

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", errs.ErrBadRequest.Wrap("failed to read file header")
	}

	contentType := http.DetectContentType(buffer[:n])

	ext, ok := detectAllowedExtension(contentType, originalName)
	if !ok {
		return "", errs.ErrBadRequest.Wrap("unsupported file type")
	}

	reader := io.MultiReader(
		bytes.NewReader(buffer[:n]),
		file,
	)

	id := uuid.New().String()

	key := fmt.Sprintf("%s.%s", id, ext)

	if err := m.storage.Save(ctx, key, reader); err != nil {
		return "", errs.ErrInternalError.Wrap("failed to save file")
	}

	return key, nil
}

func (m *mediaService) Get(ctx context.Context, key string) (io.ReadSeeker, error) {
	return m.storage.Get(ctx, key)
}

func (m *mediaService) Delete(ctx context.Context, key string) error {
	return m.storage.Delete(ctx, key)
}
