package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/storage"
)

type SubscriptionService interface {
	Create(ctx context.Context, s *models.Subscription) (*models.Subscription, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, s *models.Subscription) (*models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	//List(ctx context.Context, i interface{}) ([]models.Subscription, error)
	//TotalCost(ctx context.Context, i interface{}) (int64, error)
}

type SubscriptionServiceImpl struct {
	storage storage.SubscriptionStorage
}

func NewSubscriptionService(ss storage.SubscriptionStorage) SubscriptionService {
	return &SubscriptionServiceImpl{storage: ss}
}

func (ss *SubscriptionServiceImpl) Create(ctx context.Context, sub *models.Subscription) (*models.Subscription, error) {
	return nil, errors.New("not implemented")
}

func (ss *SubscriptionServiceImpl) Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	return nil, errors.New("not implemented")
}

func (ss *SubscriptionServiceImpl) Update(ctx context.Context, sub *models.Subscription) (*models.Subscription, error) {
	return nil, errors.New("not implemented")
}

func (ss *SubscriptionServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return errors.New("not implemented")
}

//func (s *subscriptionService) List(ctx context.Context, i interface{}) ([]models.Subscription, error) {
//	return nil, fmt.Errorf("not implemented")
//}

//func (s *subscriptionService) TotalCost(ctx context.Context, i interface{}) (int64, error) {
//	return -1, fmt.Errorf("not implemented")
//}
