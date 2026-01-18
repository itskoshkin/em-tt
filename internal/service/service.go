package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/models"
	"subscription-aggregator-service/internal/storage"
)

var (
	ErrValidationError = errors.New(fmt.Sprintf("Validation error"))
	ErrNotFound        = errors.New(fmt.Sprintf("Subscription not found"))
	ErrIES             = errors.New(fmt.Sprintf("Internal server error"))
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, s *apiModels.CreateSubscriptionRequest) (*models.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest) (*models.Subscription, error)
	UpdateSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest, sub *apiModels.UpdateSubscriptionRequest) (*models.Subscription, error)
	DeleteSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest) error
	ListSubscriptions(ctx context.Context, i interface{}) ([]models.Subscription, error)
	TotalSubscriptionsCost(ctx context.Context, i interface{}) (int64, error)
}

type SubscriptionServiceImpl struct {
	storage storage.SubscriptionStorage
}

func NewSubscriptionService(ss storage.SubscriptionStorage) SubscriptionService {
	return &SubscriptionServiceImpl{storage: ss}
}

func (ss *SubscriptionServiceImpl) CreateSubscription(ctx context.Context, req *apiModels.CreateSubscriptionRequest) (*models.Subscription, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}

	start, end, err := req.ParseDates()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}

	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      uuid.MustParse(req.UserID), //MARK: Assuming already validated above
		StartDate:   start,
		EndDate:     end,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err = ss.storage.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (ss *SubscriptionServiceImpl) GetSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest) (*models.Subscription, error) {
	uid, err := uuid.Parse(id.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid subscription UUID", ErrValidationError)
	}

	sub, err := ss.storage.GetSubscriptionByID(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return sub, nil
}

func (ss *SubscriptionServiceImpl) UpdateSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest, updated *apiModels.UpdateSubscriptionRequest) (*models.Subscription, error) {
	uid, err := uuid.Parse(id.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid subscription UUID", ErrValidationError)
	}

	if err = updated.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}

	current, err := ss.storage.GetSubscriptionByID(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	reqStart, reqEnd, clearEnd, err := updated.ParseDates()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}
	startDate := current.StartDate
	if reqStart != nil {
		startDate = *reqStart
	}
	endDate := current.EndDate
	if clearEnd {
		endDate = nil
	} else if reqEnd != nil {
		endDate = reqEnd
	}
	if endDate != nil && endDate.Before(startDate) {
		return nil, fmt.Errorf("%w: subscription end date cannot precede start date", ErrValidationError)
	}

	if updated.ServiceName != nil {
		current.ServiceName = *updated.ServiceName
	}
	if updated.Price != nil {
		current.Price = *updated.Price
	}
	//if updated.UserID != nil {
	//	current.UserID = uuid.MustParse(*updated.UserID) //MARK: Assuming already validated above; not stated if we should allow changing ID
	//}
	current.StartDate = startDate
	current.EndDate = endDate
	current.UpdatedAt = time.Now()

	if err = ss.storage.UpdateSubscriptionByID(ctx, current); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}

	return current, nil
}

func (ss *SubscriptionServiceImpl) DeleteSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest) error {
	uid, err := uuid.Parse(id.ID)
	if err != nil {
		return fmt.Errorf("%w: invalid subscription UUID", ErrValidationError)
	}

	if err = ss.storage.DeleteSubscriptionByID(ctx, uid); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		} else {
			return err
		}
	}

	return nil
}

func (ss *SubscriptionServiceImpl) ListSubscriptions(ctx context.Context, i interface{}) ([]models.Subscription, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ss *SubscriptionServiceImpl) TotalSubscriptionsCost(ctx context.Context, i interface{}) (int64, error) {
	return -1, fmt.Errorf("not implemented")
}
