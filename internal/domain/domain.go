package domain

import (
	"fmt"
	"net/http"
)

type ErrCode string

var (
	// Repository Errors
	ErrNotFound      = fmt.Errorf("not found")
	ErrAlreadyExists = fmt.Errorf("the same instance already exist")

	// Buisness Errors
	ErrMaxCountReached = fmt.Errorf("max count reached")
)

var ErrorToHTTPStatus = map[error]int{
	ErrNotFound:        http.StatusNotFound,            // 404
	ErrAlreadyExists:   http.StatusConflict,            // 409
	ErrMaxCountReached: http.StatusUnprocessableEntity, // 422
}

var errorToErrCode = map[error]ErrCode{
	ErrNotFound:        "ERR_NOT_FOUND",         // 404
	ErrAlreadyExists:   "ERR_ALREADY_EXIST",     // 409
	ErrMaxCountReached: "ERR_MAX_COUNT_REACHED", // 422
}

type Item struct {
	ID       uint
	ItemOpt1 string
	ItemOpt2 string
}
