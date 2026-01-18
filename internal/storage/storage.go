package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"subscription-aggregator-service/internal/models"
)

var ErrNotFound = errors.New("not found")

type SubscriptionStorage interface {
	CreateSubscription(ctx context.Context, s *models.Subscription) error
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	UpdateSubscriptionByID(ctx context.Context, s *models.Subscription) error
	DeleteSubscriptionByID(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, i interface{}) ([]models.Subscription, error)
}

type SubscriptionStorageImpl struct {
	db *gorm.DB
}

func NewSubscriptionsStorage(db *gorm.DB) SubscriptionStorage {
	return &SubscriptionStorageImpl{db: db}
}

func (ss *SubscriptionStorageImpl) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	return ss.db.WithContext(ctx).Create(sub).Error
}

func (ss *SubscriptionStorageImpl) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	if err := ss.db.WithContext(ctx).First(&sub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}
	return &sub, nil
}

func (ss *SubscriptionStorageImpl) UpdateSubscriptionByID(ctx context.Context, sub *models.Subscription) error {
	result := ss.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("id = ?", sub.ID).Select("service_name", "price", "user_id", "start_date", "end_date", "updated_at").
		Updates(&models.Subscription{
			ServiceName: sub.ServiceName,
			Price:       sub.Price,
			UserID:      sub.UserID,
			StartDate:   sub.StartDate,
			EndDate:     sub.EndDate,
			UpdatedAt:   time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ss *SubscriptionStorageImpl) DeleteSubscriptionByID(ctx context.Context, id uuid.UUID) error {
	result := ss.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (ss *SubscriptionStorageImpl) ListSubscriptions(ctx context.Context, i interface{}) ([]models.Subscription, error) {
	return nil, fmt.Errorf("not implemented")
}
