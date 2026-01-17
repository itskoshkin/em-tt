package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"subscription-aggregator-service/internal/models"
)

var ErrNotFound = errors.New("not found")

type SubscriptionStorage interface {
	Create(ctx context.Context, s *models.Subscription) error
	Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, s *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	//List(ctx context.Context, i interface{}) ([]models.Subscription, error)
	//TotalCost(ctx context.Context, i interface{}) (int64, error)
}

type SubscriptionStorageImpl struct {
	db *gorm.DB
}

func NewSubscriptionsStorage(db *gorm.DB) SubscriptionStorage {
	return &SubscriptionStorageImpl{db: db}
}

func (ss *SubscriptionStorageImpl) Create(ctx context.Context, sub *models.Subscription) error {
	sub.ID = uuid.New()
	return ss.db.WithContext(ctx).Create(sub).Error
}

func (ss *SubscriptionStorageImpl) Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
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

func (ss *SubscriptionStorageImpl) Update(ctx context.Context, sub *models.Subscription) error {
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

func (ss *SubscriptionStorageImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := ss.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

//func (s *SubscriptionStorageImpl) List(ctx context.Context, i interface{}) ([]models.Subscription, error) {
//	return nil, fmt.Errorf("not implemented")
//}

//func (s *SubscriptionStorageImpl) TotalCost(ctx context.Context, i interface{}) (int64, error) {
//	return -1, fmt.Errorf("not implemented")
//}
