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

	// SQL dumps (MySQL, PostgreSQL, generic)
	"application/sql": "sql",
	"text/sql":        "sql",

	// PostgreSQL dumps
	"application/x-postgresql": "dump",
	"application/postgresql":   "sql",
}

func detectAllowedExtension(contentType, originalName string) (string, bool) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(originalName), "."))

	// 1. Прямое совпадение MIME
	if e, ok := allowedTypes[contentType]; ok {
		return e, true
	}

	// 2. Частые "мусорные" MIME для дампов
	switch contentType {
	case "application/octet-stream", "text/plain":
		switch ext {
		case "sql", "dump", "db", "sqlite":
			return ext, true
		}
	}

	// 3. GZIP (очень часто .sql.gz)
	if contentType == "application/gzip" || contentType == "application/x-gzip" {
		if strings.HasSuffix(originalName, ".sql.gz") {
			return "sql.gz", true
		}
		if strings.HasSuffix(originalName, ".dump.gz") {
			return "dump.gz", true
		}
	}

	// 4. ZIP (Office + иногда дампы)
	if contentType == "application/zip" {
		switch ext {
		case "docx", "xlsx", "xls", "pptx":
			return ext, true
		case "sql", "dump":
			return ext, true
		}
	}

	// 5. Fallback по расширению (если MIME не помог)
	switch ext {
	case "sql", "dump":
		return ext, true
	}

	return "", false
}

func assignPathFromExtension(ext string) (string, bool) {

	switch ext {
	case "sql", "dump", "db", "sqlite", "sql.gz", "dump.gz":

		return "DbBackup", true
	case "docx", "xlsx", "xls", "pptx", "pdf", "odt", "ods", "odp":
		return "Docs", true
	case "jpg", "png", "webp", "gif", "bmp", "tiff", "svg":
		return "Images", true
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

	path, ok := assignPathFromExtension(ext)

	key := fmt.Sprintf("%s.%s", id, ext)
	finalKey, err := m.storage.Save(ctx, key, &path, reader)
	if err != nil {
		return "", errs.ErrInternalError.Wrap("failed to save file")
	}
	return finalKey, nil
}

func (m *mediaService) Get(ctx context.Context, key string) (io.ReadSeeker, error) {
	return m.storage.Get(ctx, key, nil)
}

func (m *mediaService) Delete(ctx context.Context, key string) error {
	return m.storage.Delete(ctx, key, nil)
}
