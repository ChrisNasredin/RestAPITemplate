package middleware

import "net/http"

type WrapperWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrapperWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.StatusCode = code
}
