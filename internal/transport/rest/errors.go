package rest

import (
	"HiveAPI/internal/domain"
	"errors"
	"net/http"
)

var (
	ErrBadRequest = errors.New("bad request")
)

var ErrorToHTTPStatus = map[error]int{
	domain.ErrNotFound:        http.StatusNotFound,            // 404
	domain.ErrAlreadyExists:   http.StatusConflict,            // 409
	domain.ErrMaxCountReached: http.StatusUnprocessableEntity, // 422
	ErrBadRequest:             http.StatusBadRequest,          //400
}
