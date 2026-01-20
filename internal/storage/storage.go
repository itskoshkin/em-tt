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
	CreateSubscription(ctx context.Context, s *models.Subscription) error
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	UpdateSubscriptionByID(ctx context.Context, s *models.Subscription) error
	DeleteSubscriptionByID(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, filter models.SubscriptionFilter) ([]models.Subscription, error)
	TotalSubscriptionsCost(ctx context.Context, filter models.SubscriptionFilter, startDate, endDate time.Time) (int64, error)
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

func (ss *SubscriptionStorageImpl) ListSubscriptions(ctx context.Context, filter models.SubscriptionFilter) ([]models.Subscription, error) {
	query := ss.db.WithContext(ctx).Model(&models.Subscription{}).Order("created_at desc, id desc")

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.ServiceName != nil {
		query = query.Where("service_name = ?", *filter.ServiceName)
	}
	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	}

	var subs []models.Subscription
	if err := query.Find(&subs).Error; err != nil {
		return nil, err
	}

	return subs, nil
}

func (ss *SubscriptionStorageImpl) TotalSubscriptionsCost(ctx context.Context, filter models.SubscriptionFilter, startDate, endDate time.Time) (int64, error) {
	query := ss.db.WithContext(ctx).Model(&models.Subscription{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.ServiceName != nil {
		query = query.Where("service_name = ?", *filter.ServiceName)
	}

	query = query.Where("start_date <= ?", endDate).
		Where("end_date IS NULL OR end_date >= ?", startDate)

	selectExpr := `
		COALESCE(SUM(
			(
				(
					EXTRACT(YEAR FROM LEAST(COALESCE(end_date, ?), ?))::int * 12 +
					EXTRACT(MONTH FROM LEAST(COALESCE(end_date, ?), ?))::int
				) -
				(
					EXTRACT(YEAR FROM GREATEST(start_date, ?))::int * 12 +
					EXTRACT(MONTH FROM GREATEST(start_date, ?))::int
				) + 1
			)::bigint * price
		), 0)
	`

	var total int64
	if err := query.Select(selectExpr,
		endDate, endDate,
		endDate, endDate,
		startDate, startDate,
	).Scan(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
