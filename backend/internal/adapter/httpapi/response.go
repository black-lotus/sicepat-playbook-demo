// Package httpapi berisi handler REST untuk dashboard monitoring.
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

type successBody struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type pageMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeData(w http.ResponseWriter, data, meta any) {
	writeJSON(w, http.StatusOK, successBody{Data: data, Meta: meta})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

// writeDomainError memetakan error domain ke status HTTP.
// Error yang tidak dikenal dicatat di log dan dikirim ke client sebagai pesan umum.
func writeDomainError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidFilter):
		writeError(w, http.StatusBadRequest, "invalid_filter", err.Error())
	default:
		logger.ErrorContext(r.Context(), "request gagal", slog.String("path", r.URL.Path), slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "terjadi kesalahan pada server")
	}
}
