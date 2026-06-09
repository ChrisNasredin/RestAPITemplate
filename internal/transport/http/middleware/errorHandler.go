package middleware

import (
	"HiveAPI/internal/transport/http/resp"
	"errors"
	"log/slog"
	"net/http"
)

type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = h(w, r)
}

type ErrResponseJSON struct {
	Message string
}

func ErrorHandler(errorMap map[error]int, log *slog.Logger) func(next APIHandler) http.Handler {
	return func(next APIHandler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := next(w, r)
			if err != nil {
				status, message := resolveHTTPStatus(err, errorMap)
				if status == http.StatusInternalServerError {
					log.Error("unknown error", slog.Any("error", err))
				}
				errResponse := ErrResponseJSON{Message: message}
				resp.ResponseJson(w, errResponse, status)
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
