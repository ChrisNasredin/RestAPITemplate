package middleware

import (
	"RestAPI/internal/lib/sl"
	"RestAPI/internal/transport/http-server"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ErrResponseJSON struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details"`
}

func ErrorHandler(errorMap map[error]int) func(next http_server.APIHandler) http.Handler {
	return func(next http_server.APIHandler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := sl.FromContext(r.Context())
			err := next(w, r)
			if err != nil {
				status, errorResp := resolveHTTPStatus(err, errorMap)
				if status == http.StatusInternalServerError {
					log.Error("internal error", slog.Any("error", err))
				}
				http_server.ResponseJson(w, errorResp, status)
			}
		})
	}
}

func resolveHTTPStatus(err error, errorMap map[error]int) (int, *ErrResponseJSON) {
	// Check Validation Error
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make(map[string]string)
		for _, fe := range ve {
			details[fe.Field()] = msgForTag(fe)
		}
		return http.StatusBadRequest, &ErrResponseJSON{
			Error:   "Validation Error",
			Details: details,
		}
	}
	// Check errorMap
	for err != nil {
		if status, exists := errorMap[err]; exists {
			return status, &ErrResponseJSON{Error: err.Error()}
		}
		err = errors.Unwrap(err)
	}
	return http.StatusInternalServerError, &ErrResponseJSON{
		Error: "Internal Server Error",
	}
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "gte":
		return "Must be greater than or equal to " + fe.Param()
	case "lte":
		return "Must be less than or equal to " + fe.Param()
	default:
		return "Invalid value"
	}
}
