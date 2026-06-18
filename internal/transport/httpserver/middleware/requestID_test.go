package middleware_test

import (
	"RestAPI/internal/lib/sl"
	"RestAPI/internal/transport/httpserver/middleware"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	// Создаем логгер для теста (в данном случае достаточно стандартного)
	logger := slog.Default()

	testCases := []struct {
		name           string
		incomingID     string
		expectRandomID bool
	}{
		{
			name:           "Generate new ID if header is missing",
			incomingID:     "",
			expectRandomID: true,
		},
		{
			name:           "Use existing ID from header",
			incomingID:     "test-request-id-123",
			expectRandomID: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			var idInHeader string
			var loggerInContext bool

			// Хендлер-заглушка для проверки прокидывания данных
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем наличие логгера в контексте
				ctxLog := sl.FromContext(r.Context())
				if ctxLog != nil {
					loggerInContext = true
				}

				// Читаем ID, который middleware уже должен был записать в ответ
				idInHeader = w.Header().Get("X-Request-ID")
				w.WriteHeader(http.StatusOK)
			})

			// Создаем middleware и оборачиваем хендлер
			mw := middleware.RequestID(logger)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.incomingID != "" {
				req.Header.Set("X-Request-Id", tt.incomingID)
			}
			rr := httptest.NewRecorder()

			// 2. Act
			mw.ServeHTTP(rr, req)

			// 3. Assert
			gotID := rr.Header().Get("X-Request-ID")

			if gotID == "" {
				t.Error("X-Request-ID header should not be empty")
			}

			if tt.expectRandomID {
				// Если ждали генерацию нового, проверяем, что он не совпадает с пустой строкой
				// и имеет достаточную длину для UUID
				if len(gotID) < 32 {
					t.Errorf("expected generated UUID, got %s", gotID)
				}
			} else {
				// Если передавали свой, проверяем, что он не изменился
				if gotID != tt.incomingID {
					t.Errorf("expected ID %s, got %s", tt.incomingID, gotID)
				}
			}

			// Проверяем, что ID в ответе совпадает с тем, что видел следующий хендлер
			if gotID != idInHeader {
				t.Errorf("ID mismatch: recorder has %s, but handler saw %s", gotID, idInHeader)
			}

			// Проверяем, что логгер был успешно прокинут в контекст
			if !loggerInContext {
				t.Error("logger was not found in request context")
			}
		})
	}
}
