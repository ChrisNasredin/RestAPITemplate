package middleware

import (
	"HiveAPI/internal/transport/rest"
	"errors"
	"log/slog"
	"net/http"
)

func ErrorHandler(errorMap map[error]int, log *slog.Logger) func(next rest.APIHandler) http.Handler {
	return func(next rest.APIHandler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := next(w, r)
			if err != nil {
				status, message := resolveHTTPStatus(err, errorMap)
				if status == http.StatusInternalServerError {
					log.Error("unknown error", slog.Any("error", err))
				}
				errResponse := rest.ErrResponseJSON{Message: message}
				rest.ResponseJson(w, errResponse, status)
			}
		})
	}
}

func resolveHTTPStatus(err error, errorMap map[error]int) (int, string) {
	for err != nil {
		if status, exists := errorMap[err]; exists {
			return status, err.Error()
		}
		err = errors.Unwrap(err)
	}
	return http.StatusInternalServerError, "Internal Server Error"

}
