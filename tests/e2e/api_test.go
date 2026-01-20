//go:build e2e

package e2e

import (
	"testing"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"subscription-aggregator-service/internal/api/controllers"
	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/service"
	"subscription-aggregator-service/internal/storage"
	"subscription-aggregator-service/tests/testutils"
)

type E2ETestSuite struct {
	suite.Suite
	container *testutils.PostgresContainer
	server    *httptest.Server
	router    *gin.Engine
	ctx       context.Context
	baseURL   string
}

func (s *E2ETestSuite) SetupSuite() {
	s.ctx = context.Background()
	gin.SetMode(gin.TestMode)

	var err error
	s.container, err = testutils.SetupPostgresContainer(s.ctx)
	require.NoError(s.T(), err, "Failed to setup postgres container")

	err = s.container.RunMigrations(s.ctx)
	require.NoError(s.T(), err, "Failed to run migrations")

	st := storage.NewSubscriptionsStorage(s.container.DB)
	svc := service.NewSubscriptionService(st)
	ctrl := controllers.NewSubscriptionController(svc)

	s.router = gin.New()
	s.router.Use(gin.Recovery())

	api := s.router.Group("/api/v1")
	{
		api.POST("/subscriptions", ctrl.CreateSubscription)
		api.GET("/subscriptions/:id", ctrl.GetSubscriptionByID)
		api.PUT("/subscriptions/:id", ctrl.UpdateSubscriptionByID)
		api.DELETE("/subscriptions/:id", ctrl.DeleteSubscriptionByID)
		api.GET("/subscriptions", ctrl.ListSubscriptions)
		api.GET("/subscriptions/total", ctrl.TotalSubscriptionsCost)
	}

	s.server = httptest.NewServer(s.router)
	s.baseURL = s.server.URL + "/api/v1"
}

func (s *E2ETestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.container != nil {
		_ = s.container.Teardown(s.ctx)
	}
}

func (s *E2ETestSuite) SetupTest() {
	err := s.container.Cleanup(s.ctx)
	require.NoError(s.T(), err)
}

func (s *E2ETestSuite) TestFullCRUDFlow() {
	userID := uuid.New().String()

	// 1. CREATE
	createReq := apiModels.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       299,
		UserID:      userID,
		StartDate:   "01-2024",
	}
	body, _ := json.Marshal(createReq)

	resp, err := http.Post(s.baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createdSub models.Subscription
	err = json.NewDecoder(resp.Body).Decode(&createdSub)
	require.NoError(s.T(), err)
	resp.Body.Close()

	assert.NotEqual(s.T(), uuid.Nil, createdSub.ID)
	assert.Equal(s.T(), "Netflix", createdSub.ServiceName)
	assert.Equal(s.T(), 299, createdSub.Price)

	// 2. READ
	resp, err = http.Get(s.baseURL + "/subscriptions/" + createdSub.ID.String())
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var retrievedSub models.Subscription
	err = json.NewDecoder(resp.Body).Decode(&retrievedSub)
	require.NoError(s.T(), err)
	resp.Body.Close()

	assert.Equal(s.T(), createdSub.ID, retrievedSub.ID)
	assert.Equal(s.T(), createdSub.ServiceName, retrievedSub.ServiceName)

	// 3. UPDATE
	newName := "Netflix Premium"
	newPrice := 499
	updateReq := apiModels.UpdateSubscriptionRequest{
		ServiceName: &newName,
		Price:       &newPrice,
	}
	body, _ = json.Marshal(updateReq)

	req, _ := http.NewRequest(http.MethodPut, s.baseURL+"/subscriptions/"+createdSub.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var updatedSub models.Subscription
	err = json.NewDecoder(resp.Body).Decode(&updatedSub)
	require.NoError(s.T(), err)
	resp.Body.Close()

	assert.Equal(s.T(), "Netflix Premium", updatedSub.ServiceName)
	assert.Equal(s.T(), 499, updatedSub.Price)

	// 4. LIST
	resp, err = http.Get(s.baseURL + "/subscriptions?user_id=" + userID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var subs []models.Subscription
	err = json.NewDecoder(resp.Body).Decode(&subs)
	require.NoError(s.T(), err)
	resp.Body.Close()

	assert.Len(s.T(), subs, 1)
	assert.Equal(s.T(), "Netflix Premium", subs[0].ServiceName)

	// 5. DELETE
	req, _ = http.NewRequest(http.MethodDelete, s.baseURL+"/subscriptions/"+createdSub.ID.String(), nil)
	resp, err = client.Do(req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// 6. Verify
	resp, err = http.Get(s.baseURL + "/subscriptions/" + createdSub.ID.String())
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func (s *E2ETestSuite) TestTotalCostCalculation() {
	userID := uuid.New().String()

	// Create multiple subscriptions
	subscriptions := []apiModels.CreateSubscriptionRequest{
		{
			ServiceName: "Netflix",
			Price:       100,
			UserID:      userID,
			StartDate:   "01-2024",
			EndDate:     strPtr("12-2024"),
		},
		{
			ServiceName: "Spotify",
			Price:       200,
			UserID:      userID,
			StartDate:   "06-2024",
		},
	}

	for _, sub := range subscriptions {
		body, _ := json.Marshal(sub)
		resp, err := http.Post(s.baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	// Calculate total cost
	url := fmt.Sprintf("%s/subscriptions/total?user_id=%s&start_date=01-2024&end_date=12-2024", s.baseURL, userID)
	resp, err := http.Get(url)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var totalResp apiModels.TotalCostResponse
	err = json.NewDecoder(resp.Body).Decode(&totalResp)
	require.NoError(s.T(), err)
	resp.Body.Close()

	// Netflix: 12 months * 100 = 1200
	// Spotify: 7 months (06-12) * 200 = 1400
	// Total = 2600
	assert.Equal(s.T(), int64(2600), totalResp.TotalCost)
}

func (s *E2ETestSuite) TestValidationErrors() {
	// Missing required fields
	createReq := apiModels.CreateSubscriptionRequest{
		Price:     299,
		UserID:    uuid.New().String(),
		StartDate: "01-2024",
	}
	body, _ := json.Marshal(createReq)

	resp, err := http.Post(s.baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Invalid UUID
	resp, err = http.Get(s.baseURL + "/subscriptions/not-a-uuid")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Invalid date format
	createReq = apiModels.CreateSubscriptionRequest{
		ServiceName: "Test",
		Price:       299,
		UserID:      uuid.New().String(),
		StartDate:   "2024-01", // Wrong format
	}
	body, _ = json.Marshal(createReq)

	resp, err = http.Post(s.baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func (s *E2ETestSuite) TestNotFound() {
	randomID := uuid.New().String()

	// GET non-existing
	resp, err := http.Get(s.baseURL + "/subscriptions/" + randomID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// DELETE non-existing
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodDelete, s.baseURL+"/subscriptions/"+randomID, nil)
	resp, err = client.Do(req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func (s *E2ETestSuite) TestListWithFilters() {
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// Create subscriptions for different users
	subs := []apiModels.CreateSubscriptionRequest{
		{ServiceName: "Netflix", Price: 100, UserID: userID1, StartDate: "01-2024"},
		{ServiceName: "Spotify", Price: 200, UserID: userID1, StartDate: "01-2024"},
		{ServiceName: "Netflix", Price: 100, UserID: userID2, StartDate: "01-2024"},
	}

	for _, sub := range subs {
		body, _ := json.Marshal(sub)
		resp, _ := http.Post(s.baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
		resp.Body.Close()
	}

	// Filter by user
	resp, _ := http.Get(s.baseURL + "/subscriptions?user_id=" + userID1)
	var result []models.Subscription
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	assert.Len(s.T(), result, 2)

	// Filter by service name
	resp, _ = http.Get(s.baseURL + "/subscriptions?service_name=Netflix")
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	assert.Len(s.T(), result, 2)

	// Filter by both
	resp, _ = http.Get(s.baseURL + "/subscriptions?user_id=" + userID1 + "&service_name=Netflix")
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	assert.Len(s.T(), result, 1)
}

func (s *E2ETestSuite) TestResponseTimes() {
	userID := uuid.New().String()

	// Create subscription
	createReq := apiModels.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       299,
		UserID:      userID,
		StartDate:   "01-2024",
	}
	body, _ := json.Marshal(createReq)

	start := time.Now()
	resp, err := http.Post(s.baseURL+"/subscriptions", "application/json", bytes.NewBuffer(body))
	require.NoError(s.T(), err)
	duration := time.Since(start)
	resp.Body.Close()

	// Response should be under 500ms for a simple operation
	assert.Less(s.T(), duration, 500*time.Millisecond, "Create operation took too long")

	// List should also be fast
	start = time.Now()
	resp, _ = http.Get(s.baseURL + "/subscriptions")
	duration = time.Since(start)
	resp.Body.Close()

	assert.Less(s.T(), duration, 500*time.Millisecond, "List operation took too long")
}

func TestE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	suite.Run(t, new(E2ETestSuite))
}

func strPtr(s string) *string {
	return &s
}
