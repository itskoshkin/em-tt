package service

import "testing"

//

import (
	"context"
	"time"

	"github.com/google/uuid"

	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/storage"
)

// MockStorage implements storage.SubscriptionStorage for testing
type MockStorage struct {
	subscriptions map[uuid.UUID]*models.Subscription
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		subscriptions: make(map[uuid.UUID]*models.Subscription),
	}
}

func (m *MockStorage) CreateSubscription(ctx context.Context, s *models.Subscription) error {
	m.subscriptions[s.ID] = s
	return nil
}

func (m *MockStorage) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	if sub, ok := m.subscriptions[id]; ok {
		return sub, nil
	}
	return nil, storage.ErrNotFound
}

func (m *MockStorage) UpdateSubscriptionByID(ctx context.Context, s *models.Subscription) error {
	if _, ok := m.subscriptions[s.ID]; !ok {
		return storage.ErrNotFound
	}
	m.subscriptions[s.ID] = s
	return nil
}

func (m *MockStorage) DeleteSubscriptionByID(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.subscriptions[id]; !ok {
		return storage.ErrNotFound
	}
	delete(m.subscriptions, id)
	return nil
}

func (m *MockStorage) ListSubscriptions(ctx context.Context, filter models.SubscriptionFilter) ([]models.Subscription, error) {
	var result []models.Subscription
	for _, sub := range m.subscriptions {
		if filter.UserID != nil && sub.UserID != *filter.UserID {
			continue
		}
		if filter.ServiceName != nil && sub.ServiceName != *filter.ServiceName {
			continue
		}
		result = append(result, *sub)
	}
	return result, nil
}

func TestCreateSubscription(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := NewSubscriptionService(mockStorage)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *apiModels.CreateSubscriptionRequest
		wantErr bool
	}{
		{
			name: "valid subscription",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test Service",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: false,
		},
		{
			name: "valid subscription with end date",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test Service",
				Price:       199,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
				EndDate:     strPtr("12-2024"),
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "zero price",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       0,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "negative price",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       -100,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "invalid user ID",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "not-a-uuid",
				StartDate:   "01-2024",
			},
			wantErr: true,
		},
		{
			name: "invalid start date format",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "2024-01",
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			req: &apiModels.CreateSubscriptionRequest{
				ServiceName: "Test",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "06-2024",
				EndDate:     strPtr("01-2024"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := svc.CreateSubscription(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateSubscription() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateSubscription() unexpected error: %v", err)
				return
			}

			if sub.ServiceName != tt.req.ServiceName {
				t.Errorf("ServiceName = %q, want %q", sub.ServiceName, tt.req.ServiceName)
			}
			if sub.Price != tt.req.Price {
				t.Errorf("Price = %d, want %d", sub.Price, tt.req.Price)
			}
		})
	}
}

func TestGetSubscriptionByID(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := NewSubscriptionService(mockStorage)
	ctx := context.Background()

	existingID := uuid.New()
	mockStorage.subscriptions[existingID] = &models.Subscription{
		ID:          existingID,
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		errType error
	}{
		{
			name:    "existing subscription",
			id:      existingID.String(),
			wantErr: false,
		},
		{
			name:    "non-existing subscription",
			id:      uuid.New().String(),
			wantErr: true,
			errType: ErrNotFound,
		},
		{
			name:    "invalid UUID",
			id:      "not-a-uuid",
			wantErr: true,
			errType: ErrValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetSubscriptionByID(ctx, apiModels.ItemByIDRequest{ID: tt.id})

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSubscriptionByID() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetSubscriptionByID() unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateSubscriptionByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		existingSub     *models.Subscription
		id              string
		req             *apiModels.UpdateSubscriptionRequest
		wantErr         bool
		wantServiceName string
		wantPrice       int
	}{
		{
			name: "update service name",
			existingSub: &models.Subscription{
				ServiceName: "Old Name",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				ServiceName: strPtr("New Name"),
			},
			wantErr:         false,
			wantServiceName: "New Name",
			wantPrice:       100,
		},
		{
			name: "update price",
			existingSub: &models.Subscription{
				ServiceName: "Test Service",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				Price: intPtr(500),
			},
			wantErr:         false,
			wantServiceName: "Test Service",
			wantPrice:       500,
		},
		{
			name: "update multiple fields",
			existingSub: &models.Subscription{
				ServiceName: "Old Name",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				ServiceName: strPtr("New Name"),
				Price:       intPtr(999),
			},
			wantErr:         false,
			wantServiceName: "New Name",
			wantPrice:       999,
		},
		{
			name: "update dates",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				StartDate: strPtr("06-2024"),
				EndDate:   strPtr("12-2024"),
			},
			wantErr:         false,
			wantServiceName: "Test",
			wantPrice:       100,
		},
		{
			name: "clear end date",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     timePtr(time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				EndDate: strPtr(""),
			},
			wantErr:         false,
			wantServiceName: "Test",
			wantPrice:       100,
		},
		{
			name:        "non-existing subscription",
			existingSub: nil,
			id:          uuid.New().String(),
			req: &apiModels.UpdateSubscriptionRequest{
				ServiceName: strPtr("New Name"),
			},
			wantErr: true,
		},
		{
			name: "invalid UUID",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			id: "not-a-uuid",
			req: &apiModels.UpdateSubscriptionRequest{
				ServiceName: strPtr("New Name"),
			},
			wantErr: true,
		},
		{
			name: "empty service name",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				ServiceName: strPtr(""),
			},
			wantErr: true,
		},
		{
			name: "zero price",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				Price: intPtr(0),
			},
			wantErr: true,
		},
		{
			name: "negative price",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				Price: intPtr(-100),
			},
			wantErr: true,
		},
		{
			name: "invalid date format",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				StartDate: strPtr("2024-01"),
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			existingSub: &models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				StartDate:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			},
			req: &apiModels.UpdateSubscriptionRequest{
				EndDate: strPtr("01-2024"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage()
			svc := NewSubscriptionService(mockStorage)

			var subID string
			if tt.existingSub != nil {
				tt.existingSub.ID = uuid.New()
				subID = tt.existingSub.ID.String()
				mockStorage.subscriptions[tt.existingSub.ID] = tt.existingSub
			} else {
				subID = tt.id
			}

			if tt.id != "" {
				subID = tt.id
			}

			result, err := svc.UpdateSubscriptionByID(ctx, apiModels.ItemByIDRequest{ID: subID}, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateSubscriptionByID() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateSubscriptionByID() unexpected error: %v", err)
				return
			}

			if result.ServiceName != tt.wantServiceName {
				t.Errorf("ServiceName = %q, want %q", result.ServiceName, tt.wantServiceName)
			}
			if result.Price != tt.wantPrice {
				t.Errorf("Price = %d, want %d", result.Price, tt.wantPrice)
			}
		})
	}
}

func TestDeleteSubscriptionByID(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := NewSubscriptionService(mockStorage)
	ctx := context.Background()

	existingID := uuid.New()
	mockStorage.subscriptions[existingID] = &models.Subscription{
		ID:          existingID,
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing subscription",
			id:      existingID.String(),
			wantErr: false,
		},
		{
			name:    "non-existing subscription",
			id:      uuid.New().String(),
			wantErr: true,
		},
		{
			name:    "invalid UUID",
			id:      "not-a-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteSubscriptionByID(ctx, apiModels.ItemByIDRequest{ID: tt.id})

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteSubscriptionByID() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteSubscriptionByID() unexpected error: %v", err)
			}
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	ctx := context.Background()

	userID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	userID2 := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	setupStorage := func() *MockStorage {
		mockStorage := NewMockStorage()

		sub1 := &models.Subscription{
			ID:          uuid.New(),
			ServiceName: "Netflix",
			Price:       100,
			UserID:      userID1,
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		sub2 := &models.Subscription{
			ID:          uuid.New(),
			ServiceName: "Spotify",
			Price:       200,
			UserID:      userID1,
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		sub3 := &models.Subscription{
			ID:          uuid.New(),
			ServiceName: "Netflix",
			Price:       100,
			UserID:      userID2,
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockStorage.subscriptions[sub1.ID] = sub1
		mockStorage.subscriptions[sub2.ID] = sub2
		mockStorage.subscriptions[sub3.ID] = sub3

		return mockStorage
	}

	tests := []struct {
		name      string
		req       apiModels.ListSubscriptionsRequest
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list all",
			req:       apiModels.ListSubscriptionsRequest{},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "filter by user ID",
			req: apiModels.ListSubscriptionsRequest{
				UserID: userID1.String(),
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "filter by service name",
			req: apiModels.ListSubscriptionsRequest{
				ServiceName: "Netflix",
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "filter by user ID and service name",
			req: apiModels.ListSubscriptionsRequest{
				UserID:      userID1.String(),
				ServiceName: "Netflix",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "no results",
			req: apiModels.ListSubscriptionsRequest{
				ServiceName: "NonExistent",
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "invalid user ID",
			req: apiModels.ListSubscriptionsRequest{
				UserID: "not-a-uuid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := setupStorage()
			svc := NewSubscriptionService(mockStorage)

			result, err := svc.ListSubscriptions(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListSubscriptions() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ListSubscriptions() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListSubscriptions() returned %d items, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestTotalSubscriptionsCost(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := NewSubscriptionService(mockStorage)
	ctx := context.Background()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Subscription 01-2024 to 12-2024, price 100
	sub1 := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Service A",
		Price:       100,
		UserID:      userID,
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     timePtr(time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)),
	}
	mockStorage.subscriptions[sub1.ID] = sub1

	// Subscription 06-2024 onwards (no end), price 200
	sub2 := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Service B",
		Price:       200,
		UserID:      userID,
		StartDate:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     nil,
	}
	mockStorage.subscriptions[sub2.ID] = sub2

	tests := []struct {
		name      string
		req       apiModels.TotalCostRequest
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "full year 2024",
			req: apiModels.TotalCostRequest{
				UserID:    userID.String(),
				StartDate: "01-2024",
				EndDate:   "12-2024",
			},
			// sub1: 12 months * 100 = 1200
			// sub2: 7 months (06-12) * 200 = 1400
			// total = 2600
			wantTotal: 2600,
			wantErr:   false,
		},
		{
			name: "first half of 2024",
			req: apiModels.TotalCostRequest{
				UserID:    userID.String(),
				StartDate: "01-2024",
				EndDate:   "06-2024",
			},
			// sub1: 6 months * 100 = 600
			// sub2: 1 month (06) * 200 = 200
			// total = 800
			wantTotal: 800,
			wantErr:   false,
		},
		{
			name: "filter by service name",
			req: apiModels.TotalCostRequest{
				UserID:      userID.String(),
				ServiceName: "Service A",
				StartDate:   "01-2024",
				EndDate:     "12-2024",
			},
			wantTotal: 1200,
			wantErr:   false,
		},
		{
			name: "invalid start date",
			req: apiModels.TotalCostRequest{
				UserID:    userID.String(),
				StartDate: "invalid",
				EndDate:   "12-2024",
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			req: apiModels.TotalCostRequest{
				UserID:    userID.String(),
				StartDate: "12-2024",
				EndDate:   "01-2024",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.TotalSubscriptionsCost(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("TotalSubscriptionsCost() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("TotalSubscriptionsCost() unexpected error: %v", err)
				return
			}

			if resp.TotalCost != tt.wantTotal {
				t.Errorf("TotalCost = %d, want %d", resp.TotalCost, tt.wantTotal)
			}
		})
	}
}

func TestCalculateSubscriptionCost(t *testing.T) {
	tests := []struct {
		name      string
		sub       models.Subscription
		startDate time.Time
		endDate   time.Time
		want      int64
	}{
		{
			name: "full overlap",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      1200, // 12 months * 100
		},
		{
			name: "subscription starts after period start",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      700, // 7 months * 100
		},
		{
			name: "subscription ends before period end",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      600, // 6 months * 100
		},
		{
			name: "no end date - uses period end",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   nil,
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			want:      600, // 6 months * 100
		},
		{
			name: "subscription outside period - before",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      0,
		},
		{
			name: "subscription outside period - after",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      0,
		},
		{
			name: "single month",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      100, // 1 month * 100
		},
		{
			name: "partial overlap at start",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      300, // 3 months (01-03) * 100
		},
		{
			name: "partial overlap at end",
			sub: models.Subscription{
				Price:     100,
				StartDate: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			want:      300, // 3 months (10-12) * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSubscriptionCost(tt.sub, tt.startDate, tt.endDate)
			if got != tt.want {
				t.Errorf("calculateSubscriptionCost() = %d, want %d", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
