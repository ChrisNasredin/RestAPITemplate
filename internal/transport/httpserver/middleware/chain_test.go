package middleware_test

import (
	"RestAPI/internal/transport/httpserver/middleware"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestChain(t *testing.T) {
	// 1. Arrange
	var trace []string

	// Вспомогательная функция для создания тестового middleware, который записывает свой вызов
	createMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	mw1 := createMiddleware("mw1")
	mw2 := createMiddleware("mw2")
	mw3 := createMiddleware("mw3")

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace = append(trace, "final_handler")
	})

	// 2. Act
	// Chain должен работать так: первое переданное middleware выполняется первым
	mw := middleware.Chain(mw1, mw2, mw3)
	handlerWithMiddleware := mw(finalHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handlerWithMiddleware.ServeHTTP(rr, req)

	// 3. Assert
	// Ожидаем порядок: mw1 -> mw2 -> mw3 -> final_handler
	expectedTrace := []string{"mw1", "mw2", "mw3", "final_handler"}

	if !reflect.DeepEqual(trace, expectedTrace) {
		t.Errorf("Execution order mismatch.\nExpected: %v\nGot:      %v", expectedTrace, trace)
	}
}
