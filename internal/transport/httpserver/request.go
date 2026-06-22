package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = NewValidator()

func HandleBody[T any](w *http.ResponseWriter, r *http.Request) (*T, error) {
	body, err := Decode[T](r.Body)
	if err != nil {
		return nil, err
	}
	err = isValid(body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}

func Decode[T any](body io.ReadCloser) (T, error) {
	var payload T
	err := json.NewDecoder(body).Decode(&payload)
	if err != nil {
		return payload, fmt.Errorf("%w: %s", ErrJSONDecode, err)
	}
	return payload, nil
}

func isValid[T any](payload T) error {

	err := validate.Struct(payload)
	if err != nil {
		return fmt.Errorf("%w: Error validate body: %w", ErrBadRequest, err)
	}
	return nil
}

// NewValidator Переопределяем конструктор валидатора, чтобы в ошибках валидации были
// корректные поля JSON, а не поля структуры
func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}
