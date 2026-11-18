package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// NewRouter menyusun semua route API beserta middleware.
func NewRouter(courierMetrics *CourierMetricHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1/courier-metrics", courierMetrics.List)

	return recoverer(logger, requestLogger(logger, mux))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// requestLogger mencatat method, path, status, dan durasi tiap request.
// Query string tidak dicatat agar parameter sensitif tidak masuk log.
func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.InfoContext(r.Context(), "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// recoverer mengubah panic dari library pihak ketiga menjadi response 500.
func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.ErrorContext(r.Context(), "panic saat menangani request", slog.Any("panic", v), slog.String("path", r.URL.Path))
				writeError(w, http.StatusInternalServerError, "internal_error", "terjadi kesalahan pada server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
