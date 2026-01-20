//go:build integration

package integration

import (
	"testing"

	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/storage"
	"subscription-aggregator-service/tests/testutils"
)

type StorageIntegrationTestSuite struct {
	suite.Suite
	container *testutils.PostgresContainer
	storage   storage.SubscriptionStorage
	ctx       context.Context
}

func (s *StorageIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	var err error
	s.container, err = testutils.SetupPostgresContainer(s.ctx)
	require.NoError(s.T(), err, "Failed to setup postgres container")

	err = s.container.RunMigrations(s.ctx)
	require.NoError(s.T(), err, "Failed to run migrations")

	s.storage = storage.NewSubscriptionsStorage(s.container.DB)
}

func (s *StorageIntegrationTestSuite) TearDownSuite() {
	if s.container != nil {
		_ = s.container.Teardown(s.ctx)
	}
}

func (s *StorageIntegrationTestSuite) SetupTest() {
	// Clean up before each test
	err := s.container.Cleanup(s.ctx)
	require.NoError(s.T(), err)
}

func (s *StorageIntegrationTestSuite) TestCreateSubscription() {
	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       299,
		UserID:      uuid.New(),
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.storage.CreateSubscription(s.ctx, sub)
	assert.NoError(s.T(), err)

	// Verify it was created
	retrieved, err := s.storage.GetSubscriptionByID(s.ctx, sub.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), sub.ServiceName, retrieved.ServiceName)
	assert.Equal(s.T(), sub.Price, retrieved.Price)
	assert.Equal(s.T(), sub.UserID, retrieved.UserID)
}

func (s *StorageIntegrationTestSuite) TestGetSubscriptionByID_NotFound() {
	_, err := s.storage.GetSubscriptionByID(s.ctx, uuid.New())
	assert.ErrorIs(s.T(), err, storage.ErrNotFound)
}

func (s *StorageIntegrationTestSuite) TestUpdateSubscription() {
	// Create subscription
	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       299,
		UserID:      uuid.New(),
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := s.storage.CreateSubscription(s.ctx, sub)
	require.NoError(s.T(), err)

	// Update
	sub.ServiceName = "Netflix Premium"
	sub.Price = 499
	err = s.storage.UpdateSubscriptionByID(s.ctx, sub)
	assert.NoError(s.T(), err)

	// Verify
	retrieved, err := s.storage.GetSubscriptionByID(s.ctx, sub.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Netflix Premium", retrieved.ServiceName)
	assert.Equal(s.T(), 499, retrieved.Price)
}

func (s *StorageIntegrationTestSuite) TestUpdateSubscription_NotFound() {
	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   time.Now(),
	}

	err := s.storage.UpdateSubscriptionByID(s.ctx, sub)
	assert.ErrorIs(s.T(), err, storage.ErrNotFound)
}

func (s *StorageIntegrationTestSuite) TestDeleteSubscription() {
	// Create subscription
	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       299,
		UserID:      uuid.New(),
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := s.storage.CreateSubscription(s.ctx, sub)
	require.NoError(s.T(), err)

	// Delete
	err = s.storage.DeleteSubscriptionByID(s.ctx, sub.ID)
	assert.NoError(s.T(), err)

	// Verify deleted (soft delete - should still return not found for normal queries)
	_, err = s.storage.GetSubscriptionByID(s.ctx, sub.ID)
	assert.ErrorIs(s.T(), err, storage.ErrNotFound)
}

func (s *StorageIntegrationTestSuite) TestDeleteSubscription_NotFound() {
	err := s.storage.DeleteSubscriptionByID(s.ctx, uuid.New())
	assert.ErrorIs(s.T(), err, storage.ErrNotFound)
}

func (s *StorageIntegrationTestSuite) TestListSubscriptions() {
	userID := uuid.New()

	// Create multiple subscriptions
	subs := []*models.Subscription{
		{
			ID:          uuid.New(),
			ServiceName: "Netflix",
			Price:       299,
			UserID:      userID,
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			ServiceName: "Spotify",
			Price:       199,
			UserID:      userID,
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			ServiceName: "Netflix",
			Price:       299,
			UserID:      uuid.New(), // Different user
			StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, sub := range subs {
		err := s.storage.CreateSubscription(s.ctx, sub)
		require.NoError(s.T(), err)
	}

	// List all
	result, err := s.storage.ListSubscriptions(s.ctx, models.SubscriptionFilter{})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 3)

	// Filter by user
	result, err = s.storage.ListSubscriptions(s.ctx, models.SubscriptionFilter{UserID: &userID})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 2)

	// Filter by service name
	serviceName := "Netflix"
	result, err = s.storage.ListSubscriptions(s.ctx, models.SubscriptionFilter{ServiceName: &serviceName})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 2)

	// Filter by both
	result, err = s.storage.ListSubscriptions(s.ctx, models.SubscriptionFilter{
		UserID:      &userID,
		ServiceName: &serviceName,
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 1)
}

func (s *StorageIntegrationTestSuite) TestConcurrentOperations() {
	// Test that concurrent operations don't cause issues
	userID := uuid.New()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			sub := &models.Subscription{
				ID:          uuid.New(),
				ServiceName: "Service",
				Price:       100 + idx,
				UserID:      userID,
				StartDate:   time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			_ = s.storage.CreateSubscription(s.ctx, sub)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all were created
	result, err := s.storage.ListSubscriptions(s.ctx, models.SubscriptionFilter{UserID: &userID})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 10)
}

func TestStorageIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(StorageIntegrationTestSuite))
}
