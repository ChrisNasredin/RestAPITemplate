package rest

import (
	"net/http"
)

type APIHandler func(w http.ResponseWriter, r *http.Request) error

type ErrResponseJSON struct {
	Message string
}

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		ResponseJson(w, ErrResponseJSON{Message: "Internal Server Error"}, http.StatusInternalServerError)
	}
}
