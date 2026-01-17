package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"subscription-aggregator-service/internal/utils/dates"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (req *CreateSubscriptionRequest) Validate() error {
	if req.ServiceName == "" { // && ∈ [A-z][0-9]?
		return fmt.Errorf("service name is required")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be above zero")
	}
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	} else if _, err := uuid.Parse(req.UserID); err != nil {
		return fmt.Errorf("user ID must be a valid UUID")
	}
	if req.StartDate == "" {
		return fmt.Errorf("start date is required") //TODO validate date format
	}
	return nil
}

func (req *CreateSubscriptionRequest) ParseDates() (time.Time, *time.Time, error) {
	start, err := dates.String2Date(req.StartDate)
	if err != nil {
		return time.Time{}, nil, err
	}

	var end time.Time
	if req.EndDate == nil {
		return start, nil, nil
	} else {
		end, err = dates.String2Date(*req.EndDate)
		if err != nil {
			return time.Time{}, nil, err
		}
	}

	if end.Before(start) {
		return time.Time{}, nil, fmt.Errorf("subscription end date cannot precede start date")
	}

	return start, &end, nil
}

type CreateSubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      string    `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	UserID      *string `json:"user_id,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (req *UpdateSubscriptionRequest) Validate() error {
	if req.ServiceName != nil && strings.TrimSpace(*req.ServiceName) == "" {
		return fmt.Errorf("service name is required")
	}
	if req.Price != nil && *req.Price <= 0 {
		return fmt.Errorf("price must be above zero")
	}
	if req.UserID != nil {
		if _, err := uuid.Parse(*req.UserID); err != nil {
			return fmt.Errorf("user ID must be a valid UUID")
		}
	}
	if req.StartDate != nil {
		if _, err := dates.String2Date(*req.StartDate); err != nil {
			return fmt.Errorf("invalid start date format")
		}
	}
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		if _, err := dates.String2Date(*req.EndDate); err != nil {
			return fmt.Errorf("invalid end date format")
		}
	}
	return nil
}

func (req *UpdateSubscriptionRequest) ParseDates() (*time.Time, *time.Time, bool, error) {
	var start *time.Time
	var end *time.Time

	if req.StartDate != nil {
		parsed, err := dates.String2Date(*req.StartDate)
		if err != nil {
			return nil, nil, false, err
		}
		start = &parsed
	}

	if req.EndDate == nil {
		return start, nil, false, nil
	}
	if strings.TrimSpace(*req.EndDate) == "" {
		return start, nil, true, nil
	}

	parsed, err := dates.String2Date(*req.EndDate)
	if err != nil {
		return nil, nil, false, err
	}
	end = &parsed

	if start != nil && end.Before(*start) {
		return nil, nil, false, fmt.Errorf("subscription end date cannot precede start date")
	}

	return start, end, false, nil
}

type ItemByIDRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}