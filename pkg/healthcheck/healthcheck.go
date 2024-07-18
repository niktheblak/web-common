package healthcheck

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/niktheblak/web-common/pkg/response"
)

func SimpleHealthCheck(logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return HealthCheck(func() error {
		return nil
	}, logger)
}

func HealthCheck(check func() error, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type status struct {
			Status string `json:"status"`
		}
		var s status
		err := check()
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "Health check error", slog.Any("error", err))
			s.Status = "error"
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			s.Status = "ok"
		}
		if err := response.Encode(w, s); err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "Error while writing HTTP response", slog.Any("error", err))
		}
	})
}
