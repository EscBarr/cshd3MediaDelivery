package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"time"

	"cshdMediaDelivery/internal/services"
	"encoding/json"
	"github.com/go-chi/chi/v5"
)

type MediaHandler struct {
	service services.MediaService
}

func NewMediaHandler(service services.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

func (h *MediaHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/upload", h.Upload)
	r.Get("/{key}", h.Get)
	r.Delete("/{key}", h.Delete)

	return r
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	key, err := h.service.Upload(r.Context(), file, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"file_id": key,
	})
}

func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "missing file key", http.StatusBadRequest)
		return
	}

	file, err := h.service.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	defer func() {
		if closer, ok := file.(io.Closer); ok {
			closer.Close()
		}
	}()

	http.ServeContent(
		w,
		r,
		filepath.Base(key),
		time.Now(),
		file,
	)
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "missing file key", http.StatusBadRequest)
		return
	}

	err := h.service.Delete(r.Context(), key)
	if err != nil {
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
