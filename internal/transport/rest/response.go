package rest

import (
	"encoding/json"
	"net/http"
)

func ResponseJson(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
