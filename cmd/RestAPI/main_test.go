package main

import (
	"RestAPI/internal/app"
	"RestAPI/internal/config"
	"RestAPI/internal/domain"
	"RestAPI/mocks"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCRUD(t *testing.T) {
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			Address: "127.0.0.1:8080",
		},
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	testCases := []struct {
		name               string
		method             string
		url                string
		body               string
		expectedBody       string
		expectedHTTPStatus int
		mockDBMethods      []func(m *mocks.MockRepositoryInterface)
		// Existing headers in response
		responseHeaders []string
	}{
		{
			name:   "Success Create Item",
			method: http.MethodPost,
			url:    "/domain",
			body: `{
				"item_opt1": "test1", 
				"item_opt2": "test1"
			}`,
			expectedBody:       `{"id":1,"item_opt1":"test1","item_opt2":"test1"}`,
			expectedHTTPStatus: 201,
			mockDBMethods: []func(m *mocks.MockRepositoryInterface){
				func(m *mocks.MockRepositoryInterface) {
					m.EXPECT().
						CreateItem(
							mock.MatchedBy(func(ctx context.Context) bool { return true }),
							&domain.Item{ItemOpt1: "test1", ItemOpt2: "test1"},
						).
						Return(&domain.Item{ID: 1, ItemOpt1: "test1", ItemOpt2: "test1"}, nil).
						Once()
				},
			},
			responseHeaders: []string{"X-Request-Id"},
		},
	}
	//// Success Create
	//mockDB.EXPECT().
	//	CreateItem(mock.Anything, &domain.Item{ItemOpt1: "test", ItemOpt2: "test"}).
	//	Return(&domain.Item{ID: 1, ItemOpt1: "test", ItemOpt2: "test"}, nil).
	//	Once()
	//// Success Update
	//mockDB.EXPECT().
	//	UpdateItem(mock.Anything, &domain.Item{ItemOpt1: "test2", ItemOpt2: "test2"}, int64(1)).
	//	Return(&domain.Item{ID: 1, ItemOpt1: "test2", ItemOpt2: "test2"}, nil).
	//	Once()
	//// Success Delete
	//mockDB.EXPECT().
	//	DeleteItem(mock.Anything, int64(1)).
	//	Return(nil).
	//	Once()
	// Success Get One
	//mockDB.EXPECT().
	//	GetItemByID(mock.MatchedBy(func(ctx context.Context) bool { return true }), int64(1)).
	//	Return(&domain.Item{ID: 1, ItemOpt1: "test1", ItemOpt2: "test1"}, nil).
	//	Once()
	////Success Get All
	//mockDB.EXPECT().
	//	GetAllItems(mock.Anything, 1, 0).
	//	Return([]*domain.Item{&domain.Item{ID: 1, ItemOpt1: "test1", ItemOpt2: "test1"}}, nil).
	//	Once()

	// httptest.NewServer()

	for _, tc := range testCases {
		// Prepare Repo Mocks
		mockDB := mocks.NewMockRepositoryInterface(t)
		testApp, err := app.New(cfg, log, mockDB)
		if err != nil {
			require.NoError(t, err)
		}
		testAppHandler := testApp.MainServerHandler()

		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, strings.NewReader(tc.expectedBody))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			// Learn mock
			for _, m := range tc.mockDBMethods {
				m(mockDB)
			}
			testAppHandler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			for _, h := range tc.responseHeaders {
				requestID := rr.Header().Get(h)
				assert.NotEmpty(t, requestID)
			}
		})
	}

}
