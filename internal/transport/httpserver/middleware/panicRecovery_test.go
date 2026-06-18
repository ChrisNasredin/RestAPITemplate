package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanicRecovery(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("some panic")
	})
	recovered := PanicRecovery(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
	rr := httptest.NewRecorder()
	expectedBody := "internal server error\n"

	recovered.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
	}
}
