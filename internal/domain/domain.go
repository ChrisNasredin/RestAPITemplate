package domain

import (
	"errors"
)

type ErrCode string

var (
	// Repository Errors
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("the same instance already exist")

	// Buisness Errors
	ErrMaxCountReached = errors.New("max count reached")
)

var errorToErrCode = map[error]ErrCode{
	ErrNotFound:        "ERR_NOT_FOUND",
	ErrAlreadyExists:   "ERR_ALREADY_EXIST",
	ErrMaxCountReached: "ERR_MAX_COUNT_REACHED",
}

type Item struct {
	ID       int64
	ItemOpt1 string
	ItemOpt2 string
}

type UpdateItemInput struct {
	ItemOpt1 *string
	ItemOpt2 *string
}
