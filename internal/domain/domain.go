package domain

import (
	"fmt"
	"net/http"
)

var (
	// Repository Errors
	ErrNotFound      = fmt.Errorf("Not found")
	ErrAlreadyExists = fmt.Errorf("Conflict")

	// Buisness Errors
	ErrMaxCountReached = fmt.Errorf("Max count reached")
)

var errorToHTTPStatus = map[error]int{
	ErrNotFound:        http.StatusNotFound,            // 404
	ErrAlreadyExists:   http.StatusConflict,            // 409
	ErrMaxCountReached: http.StatusUnprocessableEntity, // 422
}

type Item struct {
	ID       uint
	ItemOpt1 string
	ItemOpt2 string
}
