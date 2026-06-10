package http_server

import (
	"log/slog"
	"net/http"
)

type APIHandler func(w http.ResponseWriter, r *http.Request) error

type ErrResponseJSON struct {
	Message string
}

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("We are call method ServeHTTP")
	err := h(w, r)
	if err != nil {
		ResponseJson(w, ErrResponseJSON{Message: "Internal Server Error"}, http.StatusInternalServerError)
	}
}
