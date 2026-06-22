package middleware

import (
	"RestAPI/internal/lib/sl"
	"RestAPI/internal/transport/httpserver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ErrResponseJSON struct {
	Error   string `json:"error"`
	Details any    `json:"details"`
}

var (
	// Validation JSON Errors
	validationErr validator.ValidationErrors
	// JSON Decoding Errors
	JSONSyntaxErr         *json.SyntaxError
	JSONUnmarshTypeErr    *json.UnmarshalTypeError
	JSONInvalidUnmarshErr *json.InvalidUnmarshalError
	JSONEmpty             = io.EOF
)

func ErrorHandler(errorMap map[error]int) func(next httpserver.APIHandler) http.Handler {
	return func(next httpserver.APIHandler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := sl.FromContext(r.Context())
			err := next(w, r)
			if err != nil {
				status, errorResp := resolveHTTPStatus(err, errorMap)
				if status == http.StatusInternalServerError {
					log.Error("internal error", slog.Any("error", err))
				}
				httpserver.ResponseJson(w, errorResp, status)
			}
		})
	}
}

func resolveHTTPStatus(err error, errorMap map[error]int) (int, *ErrResponseJSON) {
	// TODO: Убрать в функцию
	// Check Validation Error
	if errors.As(err, &validationErr) {
		details := make(map[string]string)
		for _, fe := range validationErr {
			details[fe.Field()] = msgForTag(fe)
		}
		return http.StatusUnprocessableEntity, &ErrResponseJSON{
			Error:   "Validation Error",
			Details: details,
		}
	}

	if details, isValidationErr := handleDecodeError(err); isValidationErr {
		return http.StatusBadRequest, &ErrResponseJSON{
			Error:   "JSON Decode Error",
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

func handleDecodeError(err error) (string, bool) {
	if errors.Is(err, JSONEmpty) {
		return "Empty json body", true
	}
	if errors.As(err, &JSONSyntaxErr) {
		return fmt.Sprintf("JSON syntax error: %s", JSONSyntaxErr.Error()), true
	}
	if errors.As(err, &JSONUnmarshTypeErr) {
		return fmt.Sprintf(
			"JSON type error: field %q has type %s, require type %s",
			JSONUnmarshTypeErr.Field,
			JSONUnmarshTypeErr.Value,
			JSONUnmarshTypeErr.Type,
		), true
	}
	if errors.As(err, &JSONInvalidUnmarshErr) {
		// Ошибка декодинга, но должна пятисотить и логироваться
		return "", false
	}
	return "", false
}
