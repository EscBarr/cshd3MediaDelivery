package handlers

import (
	"cshdMediaDelivery/internal/lib/errs"
	response "cshdMediaDelivery/internal/responce"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"cshdMediaDelivery/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type MediaHandler struct {
	service services.MediaService
}

func NewMediaHandler(service services.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	file, header, err := r.FormFile("file")
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(
			errs.ErrBadRequest.Wrap("failed to read multipart file"),
		))
		return
	}
	defer file.Close()

	key, err := h.service.Upload(ctx, file, header.Filename)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusCreated,
		Data: map[string]string{
			"file_id": key,
		},
	})
}

func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	key := chi.URLParam(r, "key")
	if key == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(
			errs.ErrBadRequest.Wrap("missing file key"),
		))
		return
	}

	file, err := h.service.Get(ctx, key)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrNotFound))
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
	ctx := r.Context()

	key := chi.URLParam(r, "key")
	if key == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorApiResponse(
			errs.ErrBadRequest.Wrap("missing file key"),
		))
		return
	}

	err := h.service.Delete(ctx, key)
	if err != nil {
		if apiErr, ok := errs.IsApiError(err); ok {
			render.Status(r, apiErr.HttpCode)
			render.JSON(w, r, response.ErrorApiResponse(apiErr))
			return
		}

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorApiResponse(errs.ErrInternalError))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response.ApiResponse{
		Status:     response.StatusOK,
		StatusCode: http.StatusOK,
		Message:    "file deleted successfully",
	})
}
