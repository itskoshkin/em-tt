package controllers

import "testing"

//

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/service"
)

// MockSubscriptionService implements service.SubscriptionService for testing
type MockSubscriptionService struct {
	subscriptions map[uuid.UUID]*models.Subscription
}

func NewMockService() *MockSubscriptionService {
	return &MockSubscriptionService{
		subscriptions: make(map[uuid.UUID]*models.Subscription),
	}
}

func (m *MockSubscriptionService) CreateSubscription(ctx context.Context, req *apiModels.CreateSubscriptionRequest) (*models.Subscription, error) {
	if err := req.Validate(); err != nil {
		return nil, service.ErrValidationError
	}

	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      uuid.MustParse(req.UserID),
		StartDate:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.subscriptions[sub.ID] = sub
	return sub, nil
}

func (m *MockSubscriptionService) GetSubscriptionByID(ctx context.Context, req apiModels.ItemByIDRequest) (*models.Subscription, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, service.ErrValidationError
	}

	if sub, ok := m.subscriptions[id]; ok {
		return sub, nil
	}
	return nil, service.ErrNotFound
}

func (m *MockSubscriptionService) UpdateSubscriptionByID(ctx context.Context, req apiModels.ItemByIDRequest, update *apiModels.UpdateSubscriptionRequest) (*models.Subscription, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, service.ErrValidationError
	}

	sub, ok := m.subscriptions[id]
	if !ok {
		return nil, service.ErrNotFound
	}

	if update.ServiceName != nil {
		sub.ServiceName = *update.ServiceName
	}
	if update.Price != nil {
		sub.Price = *update.Price
	}
	sub.UpdatedAt = time.Now()

	return sub, nil
}

func (m *MockSubscriptionService) DeleteSubscriptionByID(ctx context.Context, req apiModels.ItemByIDRequest) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return service.ErrValidationError
	}

	if _, ok := m.subscriptions[id]; !ok {
		return service.ErrNotFound
	}
	delete(m.subscriptions, id)
	return nil
}

func (m *MockSubscriptionService) ListSubscriptions(ctx context.Context, req apiModels.ListSubscriptionsRequest) ([]models.Subscription, error) {
	var result []models.Subscription
	for _, sub := range m.subscriptions {
		result = append(result, *sub)
	}
	return result, nil
}

func (m *MockSubscriptionService) TotalSubscriptionsCost(ctx context.Context, req apiModels.TotalCostRequest) (*apiModels.TotalCostResponse, error) {
	return &apiModels.TotalCostResponse{TotalCost: 1000}, nil
}

func setupRouter(ctrl *SubscriptionController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/subscriptions", ctrl.CreateSubscription)
	r.GET("/subscriptions/:id", ctrl.GetSubscriptionByID)
	r.PUT("/subscriptions/:id", ctrl.UpdateSubscriptionByID)
	r.DELETE("/subscriptions/:id", ctrl.DeleteSubscriptionByID)
	r.GET("/subscriptions", ctrl.ListSubscriptions)
	r.GET("/subscriptions/total", ctrl.TotalSubscriptionsCost)

	return r
}

func TestCreateSubscriptionHandler(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	tests := []struct {
		name           string
		body           interface{}
		wantStatusCode int
	}{
		{
			name: "valid request",
			body: apiModels.CreateSubscriptionRequest{
				ServiceName: "Netflix",
				Price:       299,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			body:           "not json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing required field",
			body: apiModels.CreateSubscriptionRequest{
				Price:     299,
				UserID:    "550e8400-e29b-41d4-a716-446655440000",
				StartDate: "01-2024",
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("CreateSubscription() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestGetSubscriptionByIDHandler(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	// Pre-create a subscription
	existingID := uuid.New()
	mockService.subscriptions[existingID] = &models.Subscription{
		ID:          existingID,
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	tests := []struct {
		name           string
		id             string
		wantStatusCode int
	}{
		{
			name:           "existing subscription",
			id:             existingID.String(),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "non-existing subscription",
			id:             uuid.New().String(),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			id:             "not-a-uuid",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+tt.id, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("GetSubscriptionByID() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestUpdateSubscriptionByIDHandler(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	existingID := uuid.New()
	mockService.subscriptions[existingID] = &models.Subscription{
		ID:          existingID,
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	tests := []struct {
		name           string
		id             string
		body           interface{}
		wantStatusCode int
	}{
		{
			name: "valid update",
			id:   existingID.String(),
			body: apiModels.UpdateSubscriptionRequest{
				ServiceName: strPtr("Updated Name"),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "non-existing subscription",
			id:             uuid.New().String(),
			body:           apiModels.UpdateSubscriptionRequest{ServiceName: strPtr("Test")},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			id:             "not-a-uuid",
			body:           apiModels.UpdateSubscriptionRequest{ServiceName: strPtr("Test")},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			id:             existingID.String(),
			body:           "not json",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("UpdateSubscriptionByID() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestDeleteSubscriptionByIDHandler(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	existingID := uuid.New()
	mockService.subscriptions[existingID] = &models.Subscription{
		ID:          existingID,
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	tests := []struct {
		name           string
		id             string
		wantStatusCode int
	}{
		{
			name:           "existing subscription",
			id:             existingID.String(),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "non-existing subscription",
			id:             uuid.New().String(),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "invalid UUID",
			id:             "not-a-uuid",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+tt.id, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("DeleteSubscriptionByID() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestListSubscriptionsHandler(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	// Pre-create subscriptions
	for i := 0; i < 3; i++ {
		id := uuid.New()
		mockService.subscriptions[id] = &models.Subscription{
			ID:          id,
			ServiceName: "Test",
			Price:       100,
			UserID:      uuid.New(),
			StartDate:   time.Now(),
		}
	}

	tests := []struct {
		name           string
		query          string
		wantStatusCode int
	}{
		{
			name:           "list all",
			query:          "",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "with user_id filter",
			query:          "?user_id=550e8400-e29b-41d4-a716-446655440000",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "with service_name filter",
			query:          "?service_name=Netflix",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/subscriptions"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("ListSubscriptions() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestTotalSubscriptionsCostHandler(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	tests := []struct {
		name           string
		query          string
		wantStatusCode int
	}{
		{
			name:           "valid request",
			query:          "?start_date=01-2024&end_date=12-2024",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "with user_id",
			query:          "?start_date=01-2024&end_date=12-2024&user_id=550e8400-e29b-41d4-a716-446655440000",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing start_date",
			query:          "?end_date=12-2024",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing end_date",
			query:          "?start_date=01-2024",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/subscriptions/total"+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("TotalSubscriptionsCost() status = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestResponseFormat(t *testing.T) {
	mockService := NewMockService()
	ctrl := NewSubscriptionController(mockService)
	router := setupRouter(ctrl)

	// Create subscription and verify response format
	body, _ := json.Marshal(apiModels.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       299,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2024",
	})

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("CreateSubscription() status = %d, want %d", w.Code, http.StatusCreated)
	}

	var response models.Subscription
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.ID == uuid.Nil {
		t.Error("response ID should not be nil")
	}
	if response.ServiceName != "Netflix" {
		t.Errorf("ServiceName = %q, want %q", response.ServiceName, "Netflix")
	}
	if response.Price != 299 {
		t.Errorf("Price = %d, want %d", response.Price, 299)
	}
}

func strPtr(s string) *string {
	return &s
}
