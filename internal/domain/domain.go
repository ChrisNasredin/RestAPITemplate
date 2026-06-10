package domain

import (
	"fmt"
)

type ErrCode string

var (
	// Repository Errors
	ErrNotFound      = fmt.Errorf("not found")
	ErrAlreadyExists = fmt.Errorf("the same instance already exist")

	// Buisness Errors
	ErrMaxCountReached = fmt.Errorf("max count reached")
)

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
