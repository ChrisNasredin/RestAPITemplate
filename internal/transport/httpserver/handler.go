package httpserver

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
	_ = h(w, r)
}
