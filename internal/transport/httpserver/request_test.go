package httpserver_test

import (
	http_server "RestAPI/internal/transport/httpserver"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHandleBody(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		body        string
		expected    *http_server.CreateItemRequest
		expectedErr error
	}{
		// TODO: Эти тесты нужно переделать, они должны возвращать ошибки декодинга и валидации теперь
		//{name: "Invalid JSON body", body: `{item_opt1: "string", item_opt2: "string"}`, expected: nil, expectedErr: http_server.ErrBadRequest},
		//{name: "No required field", body: `{"item_opt1": "string"}`, expected: nil, expectedErr: http_server.ErrBadRequest},
		{name: "Success", body: `{"item_opt1": "string", "item_opt2": "string"}`, expected: &http_server.CreateItemRequest{ItemOpt1: "string", ItemOpt2: "string"}, expectedErr: nil},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			got, err := http_server.HandleBody[http_server.CreateItemRequest](nil, req)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("HandleBody() error = %v, expected to wrap %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("HandleBody() error = %v, expected to wrap nil", err)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("HandleBody() got = %+v, want %+v", got, tt.expected)
			}
		})
	}
}
