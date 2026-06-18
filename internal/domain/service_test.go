package domain_test

import (
	"RestAPI/internal/domain"
	"RestAPI/mocks"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ErrDatabaseFailure = errors.New("DB failure")

func TestService_GetItem(t *testing.T) {

	testcases := []struct {
		name           string
		expectedError  error
		itemID         int64
		expectedResult *domain.Item
		mock           func(m *mocks.MockRepositoryInterface, ctx context.Context, id int64)
	}{
		{
			name:          "success",
			expectedError: nil,
			itemID:        1,
			mock: func(m *mocks.MockRepositoryInterface, ctx context.Context, id int64) {
				m.EXPECT().GetItemByID(ctx, id).Return(&domain.Item{ID: 1, ItemOpt1: "test", ItemOpt2: "test"}, nil).Once()
			},
			expectedResult: &domain.Item{ID: 1, ItemOpt1: "test", ItemOpt2: "test"},
		},
		{
			name:          "DB failure",
			expectedError: ErrDatabaseFailure,
			itemID:        1,
			mock: func(m *mocks.MockRepositoryInterface, ctx context.Context, id int64) {
				m.EXPECT().GetItemByID(ctx, id).Return(nil, ErrDatabaseFailure).Once()
			},
			expectedResult: nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockRepo := mocks.NewMockRepositoryInterface(t)
			svc := domain.NewService(mockRepo)
			ctx := context.Background()
			tc.mock(mockRepo, ctx, tc.itemID)
			// Act
			result, err := svc.GetItem(ctx, tc.itemID)
			// Assert
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestService_CreateItem(t *testing.T) {

	testcases := []struct {
		name           string
		expectedError  error
		expectedResult *domain.Item
		mock           func(m *mocks.MockRepositoryInterface, ctx context.Context)
	}{
		{
			name:          "success",
			expectedError: nil,
			mock: func(m *mocks.MockRepositoryInterface, ctx context.Context) {
				m.EXPECT().CreateItem(ctx, &domain.Item{ItemOpt1: "test", ItemOpt2: "test"}).
					Return(&domain.Item{ID: 1, ItemOpt1: "test", ItemOpt2: "test"}, nil).Once()
			},
			expectedResult: &domain.Item{ID: 1, ItemOpt1: "test", ItemOpt2: "test"},
		},
		{
			name:          "DB failure",
			expectedError: ErrDatabaseFailure,
			mock: func(m *mocks.MockRepositoryInterface, ctx context.Context) {
				m.EXPECT().CreateItem(ctx, &domain.Item{ItemOpt1: "test", ItemOpt2: "test"}).
					Return(nil, ErrDatabaseFailure).Once()
			},
			expectedResult: nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockRepo := mocks.NewMockRepositoryInterface(t)
			svc := domain.NewService(mockRepo)
			ctx := context.Background()
			tc.mock(mockRepo, ctx)
			// Act
			result, err := svc.CreateItem(ctx, &domain.Item{ItemOpt1: "test", ItemOpt2: "test"})
			// Assert
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
