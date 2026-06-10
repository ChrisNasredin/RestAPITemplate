package middleware

import (
	"HiveAPI/internal/lib/sl"
	"HiveAPI/internal/transport/http-server"
	"errors"
	"log/slog"
	"net/http"
)

func ErrorHandler(errorMap map[error]int) func(next http_server.APIHandler) http.Handler {
	return func(next http_server.APIHandler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := sl.FromContext(r.Context())
			err := next(w, r)
			if err != nil {
				status, message := resolveHTTPStatus(err, errorMap)
				if status == http.StatusInternalServerError {
					log.Error("internal error", slog.Any("error", err))
				}
				errResponse := http_server.ErrResponseJSON{Message: message}
				http_server.ResponseJson(w, errResponse, status)
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
