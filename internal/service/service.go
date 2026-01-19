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
	"subscription-aggregator-service/internal/utils/dates"
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
	ListSubscriptions(ctx context.Context, req apiModels.ListSubscriptionsRequest) ([]models.Subscription, error)
	TotalSubscriptionsCost(ctx context.Context, req apiModels.TotalCostRequest) (*apiModels.TotalCostResponse, error)
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

func (ss *SubscriptionServiceImpl) ListSubscriptions(ctx context.Context, req apiModels.ListSubscriptionsRequest) ([]models.Subscription, error) {
	filter := models.SubscriptionFilter{}

	if req.UserID != "" {
		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid user ID", ErrValidationError)
		}
		filter.UserID = &uid
	}
	if req.ServiceName != "" {
		filter.ServiceName = &req.ServiceName
	}

	return ss.storage.ListSubscriptions(ctx, filter)
}

func (ss *SubscriptionServiceImpl) TotalSubscriptionsCost(ctx context.Context, req apiModels.TotalCostRequest) (*apiModels.TotalCostResponse, error) {
	startDate, err := dates.String2Date(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start date", ErrValidationError)
	}
	endDate, err := dates.String2Date(req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid end date", ErrValidationError)
	}
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("%w: end date cannot precede start date", ErrValidationError)
	}

	filter := models.SubscriptionFilter{}
	if req.UserID != "" {
		var uid uuid.UUID
		uid, err = uuid.Parse(req.UserID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid user ID", ErrValidationError)
		}
		filter.UserID = &uid
	}
	if req.ServiceName != "" {
		filter.ServiceName = &req.ServiceName
	}

	subs, err := ss.storage.ListSubscriptions(ctx, filter)
	if err != nil {
		return nil, err
	}

	var totalCost int64
	for _, sub := range subs {
		totalCost += calculateSubscriptionCost(sub, startDate, endDate)
	}

	return &apiModels.TotalCostResponse{TotalCost: totalCost}, nil
}

func calculateSubscriptionCost(sub models.Subscription, startDate, endDate time.Time) int64 {
	start := startDate
	if sub.StartDate.After(endDate) {
		start = sub.StartDate
	}
	end := endDate
	if sub.EndDate != nil && sub.EndDate.Before(endDate) {
		end = *sub.EndDate
	}
	if end.Before(start) {
		return 0
	}

	y1, m1, _ := start.Date()
	y2, m2, _ := end.Date()
	months := (y2-y1)*12 + int(m2-m1) + 1
	if months < 0 {
		return 0
	} else {
		return int64(months) * int64(sub.Price)
	}
}
