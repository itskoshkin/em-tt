package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
		slog.Warn("failed to validate subscription payload", "error", err)
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}

	start, end, err := req.ParseDates()
	if err != nil {
		slog.Warn("failed to validate subscription dates", "error", err)
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}

	sub := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      uuid.MustParse(req.UserID), // Assuming already validated above
		StartDate:   start,
		EndDate:     end,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err = ss.storage.CreateSubscription(ctx, sub); err != nil {
		slog.Error("failed to create subscription in database", "error", err)
		return nil, err
	}

	slog.Info("subscription created", "id", sub.ID, "user_id", sub.UserID)
	return sub, nil
}

func (ss *SubscriptionServiceImpl) GetSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest) (*models.Subscription, error) {
	uid, err := uuid.Parse(id.ID)
	if err != nil {
		slog.Warn("failed to validate subscription id", "error", err)
		return nil, fmt.Errorf("%w: invalid subscription UUID", ErrValidationError)
	}

	sub, err := ss.storage.GetSubscriptionByID(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			slog.Warn("requested subscription not found", "error", err)
			return nil, ErrNotFound
		} else {
			slog.Error("failed to get subscription from database", "error", err)
			return nil, err
		}
	}

	slog.Debug("subscription retrieved", "id", uid)
	return sub, nil
}

func (ss *SubscriptionServiceImpl) UpdateSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest, updated *apiModels.UpdateSubscriptionRequest) (*models.Subscription, error) {
	uid, err := uuid.Parse(id.ID)
	if err != nil {
		slog.Warn("failed to validate subscription id", "error", err)
		return nil, fmt.Errorf("%w: invalid subscription UUID", ErrValidationError)
	}

	if err = updated.Validate(); err != nil {
		slog.Warn("failed to validate subscription payload", "error", err)
		return nil, fmt.Errorf("%w: %s", ErrValidationError, err.Error())
	}

	current, err := ss.storage.GetSubscriptionByID(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			slog.Warn("requested subscription not found", "error", err)
			return nil, ErrNotFound
		} else {
			slog.Error("failed to get subscription from database", "error", err)
			return nil, err
		}
	}

	reqStart, reqEnd, clearEnd, err := updated.ParseDates()
	if err != nil {
		slog.Warn("failed to validate subscription dates", "error", err)
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
		slog.Warn("failed to validate subscription dates", "error", err)
		return nil, fmt.Errorf("%w: subscription end date cannot precede start date", ErrValidationError)
	}

	if updated.ServiceName != nil {
		current.ServiceName = *updated.ServiceName
	}
	if updated.Price != nil {
		current.Price = *updated.Price
	}
	current.StartDate = startDate
	current.EndDate = endDate
	current.UpdatedAt = time.Now()

	if err = ss.storage.UpdateSubscriptionByID(ctx, current); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			slog.Warn("requested subscription not found", "error", err)
			return nil, ErrNotFound
		} else {
			slog.Error("failed to update subscription in database", "error", err)
			return nil, err
		}
	}

	slog.Info("subscription updated", "id", uid)
	return current, nil
}

func (ss *SubscriptionServiceImpl) DeleteSubscriptionByID(ctx context.Context, id apiModels.ItemByIDRequest) error {
	uid, err := uuid.Parse(id.ID)
	if err != nil {
		slog.Warn("failed to validate subscription id", "error", err)
		return fmt.Errorf("%w: invalid subscription UUID", ErrValidationError)
	}

	if err = ss.storage.DeleteSubscriptionByID(ctx, uid); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			slog.Warn("requested subscription not found", "error", err)
			return ErrNotFound
		} else {
			slog.Error("failed to delete subscription in database", "error", err)
			return err
		}
	}

	slog.Info("subscription deleted", "id", uid)
	return nil
}

func (ss *SubscriptionServiceImpl) ListSubscriptions(ctx context.Context, req apiModels.ListSubscriptionsRequest) ([]models.Subscription, error) {
	filter := models.SubscriptionFilter{}

	if req.UserID != "" {
		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			slog.Warn("failed to validate user ID", "error", err)
			return nil, fmt.Errorf("%w: invalid user ID", ErrValidationError)
		}
		filter.UserID = &uid
	}
	if req.ServiceName != "" {
		filter.ServiceName = &req.ServiceName
	}
	if req.Limit != nil {
		if *req.Limit <= 0 {
			slog.Warn("failed to validate limit", "limit", *req.Limit)
			return nil, fmt.Errorf("%w: invalid limit", ErrValidationError)
		}
		filter.Limit = req.Limit
	}
	if req.Offset != nil {
		if *req.Offset < 0 {
			slog.Warn("failed to validate offset", "offset", *req.Offset)
			return nil, fmt.Errorf("%w: invalid offset", ErrValidationError)
		}
		filter.Offset = req.Offset
	}

	list, err := ss.storage.ListSubscriptions(ctx, filter)
	if err != nil {
		slog.Error("failed to list subscriptions from database", "error", err)
		return nil, err
	}

	slog.Debug("subscriptions list retrieved", "id_filter", filter.UserID, "service_filter", filter.ServiceName, "limit", filter.Limit, "offset", filter.Offset)
	return list, nil
}

func (ss *SubscriptionServiceImpl) TotalSubscriptionsCost(ctx context.Context, req apiModels.TotalCostRequest) (*apiModels.TotalCostResponse, error) {
	startDate, err := dates.String2Date(req.StartDate)
	if err != nil {
		slog.Warn("failed to validate subscription dates", "error", err)
		return nil, fmt.Errorf("%w: invalid start date", ErrValidationError)
	}
	endDate, err := dates.String2Date(req.EndDate)
	if err != nil {
		slog.Warn("failed to validate subscription dates", "error", err)
		return nil, fmt.Errorf("%w: invalid end date", ErrValidationError)
	}
	if endDate.Before(startDate) {
		slog.Warn("failed to validate subscription dates", "error", err)
		return nil, fmt.Errorf("%w: end date cannot precede start date", ErrValidationError)
	}

	filter := models.SubscriptionFilter{}
	if req.UserID != "" {
		var uid uuid.UUID
		uid, err = uuid.Parse(req.UserID)
		if err != nil {
			slog.Warn("failed to validate user ID", "error", err)
			return nil, fmt.Errorf("%w: invalid user ID", ErrValidationError)
		}
		filter.UserID = &uid
	}
	if req.ServiceName != "" {
		filter.ServiceName = &req.ServiceName
	}

	subs, err := ss.storage.ListSubscriptions(ctx, filter)
	if err != nil {
		slog.Error("failed to list subscriptions from database", "error", err)
		return nil, err
	}

	var totalCost int64
	for _, sub := range subs {
		totalCost += calculateSubscriptionCost(sub, startDate, endDate)
	}

	slog.Info("calculated total cost", "user_id", req.UserID, "total", totalCost, "start", startDate.Format("01-2006"), "end", endDate.Format("01-2006"))
	return &apiModels.TotalCostResponse{TotalCost: totalCost}, nil
}

func calculateSubscriptionCost(sub models.Subscription, startDate, endDate time.Time) int64 {
	start := startDate
	if sub.StartDate.After(startDate) {
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
