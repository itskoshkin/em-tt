//go:build e2e

package e2e

import (
	"testing"

	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"subscription-aggregator-service/internal/api/controllers"
	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/service"
	"subscription-aggregator-service/internal/storage"
	"subscription-aggregator-service/tests/testutils"
)

func TestSmoke_ServiceStarts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke tests in short mode")
	}

	ctx := context.Background()
	gin.SetMode(gin.TestMode)

	// Setup container
	container, err := testutils.SetupPostgresContainer(ctx)
	require.NoError(t, err, "Service should be able to connect to database")
	defer container.Teardown(ctx)

	err = container.RunMigrations(ctx)
	require.NoError(t, err, "Migrations should run successfully")

	// Setup application
	st := storage.NewSubscriptionsStorage(container.DB)
	svc := service.NewSubscriptionService(st)
	ctrl := controllers.NewSubscriptionController(svc)

	router := gin.New()
	router.Use(gin.Recovery())
	api := router.Group("/api/v1")
	{
		api.POST("/subscriptions", ctrl.CreateSubscription)
		api.GET("/subscriptions/:id", ctrl.GetSubscriptionByID)
		api.GET("/subscriptions", ctrl.ListSubscriptions)
		api.GET("/subscriptions/total", ctrl.TotalSubscriptionsCost)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	baseURL := server.URL + "/api/v1"

	t.Run("List endpoint responds", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/subscriptions")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Create endpoint responds", func(t *testing.T) {
		req := apiModels.CreateSubscriptionRequest{
			ServiceName: "Test",
			Price:       100,
			UserID:      uuid.New().String(),
			StartDate:   "01-2024",
		}
		body, _ := json.Marshal(req)

		resp, err := http.Post(baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Get endpoint responds with 404 for non-existing", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/subscriptions/" + uuid.New().String())
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Total cost endpoint responds", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/subscriptions/total?start_date=01-2024&end_date=12-2024")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Invalid request returns 400", func(t *testing.T) {
		resp, err := http.Post(baseURL+"/subscriptions", "application/json", bytes.NewBufferString("invalid"))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestSmoke_HandlersRespond(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockService{}
	ctrl := controllers.NewSubscriptionController(mockSvc)

	router := gin.New()
	api := router.Group("/api/v1")
	{
		api.POST("/subscriptions", ctrl.CreateSubscription)
		api.GET("/subscriptions/:id", ctrl.GetSubscriptionByID)
		api.GET("/subscriptions", ctrl.ListSubscriptions)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	baseURL := server.URL + "/api/v1"

	// Just verify endpoints are reachable
	endpoints := []string{
		"/subscriptions",
	}

	for _, endpoint := range endpoints {
		t.Run("GET "+endpoint, func(t *testing.T) {
			resp, err := http.Get(baseURL + endpoint)
			require.NoError(t, err)
			resp.Body.Close()
			assert.NotEqual(t, 0, resp.StatusCode)
		})
	}
}

// Mock service for lightweight smoke tests
type mockService struct{}

func (m *mockService) CreateSubscription(ctx context.Context, req *apiModels.CreateSubscriptionRequest) (*models.Subscription, error) {
	return nil, service.ErrValidationError
}

func (m *mockService) GetSubscriptionByID(ctx context.Context, req apiModels.ItemByIDRequest) (*models.Subscription, error) {
	return nil, service.ErrNotFound
}

func (m *mockService) UpdateSubscriptionByID(ctx context.Context, req apiModels.ItemByIDRequest, update *apiModels.UpdateSubscriptionRequest) (*models.Subscription, error) {
	return nil, service.ErrNotFound
}

func (m *mockService) DeleteSubscriptionByID(ctx context.Context, req apiModels.ItemByIDRequest) error {
	return service.ErrNotFound
}

func (m *mockService) ListSubscriptions(ctx context.Context, req apiModels.ListSubscriptionsRequest) ([]models.Subscription, error) {
	return []models.Subscription{}, nil
}

func (m *mockService) TotalSubscriptionsCost(ctx context.Context, req apiModels.TotalCostRequest) (*apiModels.TotalCostResponse, error) {
	return &apiModels.TotalCostResponse{TotalCost: 0}, nil
}
