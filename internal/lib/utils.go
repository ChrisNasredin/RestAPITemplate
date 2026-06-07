package lib

import (
	"errors"
	"log/slog"
	"net/http"
)

func ResolveHTTPStatus(err error, errorMap map[error]int, log *slog.Logger) (int, string) {
	originalErr := err
	log.Debug("", slog.Any("error", originalErr))
	for err != nil {
		if status, exists := errorMap[err]; exists {
			return status, err.Error()
		}
		err = errors.Unwrap(err)
	}
	log.Warn("Unhandled Error", slog.Any("error", originalErr))
	return http.StatusInternalServerError, "Internal Server Error"

}
