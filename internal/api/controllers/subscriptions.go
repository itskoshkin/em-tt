package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	apiModels "subscription-aggregator-service/internal/api/models"
	_ "subscription-aggregator-service/internal/models" // Make visible for swaggo/swag tool
	"subscription-aggregator-service/internal/service"
)

type SubscriptionController struct {
	subscriptionService service.SubscriptionService
}

func NewSubscriptionController(ss service.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{subscriptionService: ss}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Adds a new subscription record to the database with given details
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body apiModels.CreateSubscriptionRequest true "New subscription details"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions [post]
func (ctrl *SubscriptionController) CreateSubscription(ctx *gin.Context) {
	var req apiModels.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: apiModels.ErrBadJSON.Error()})
		return
	}

	sub, err := ctrl.subscriptionService.CreateSubscription(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: service.ErrIES.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, sub)
}

// GetSubscriptionByID godoc
// @Summary Get a subscription by ID
// @Description Returns a single subscription record by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription UUID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} apiModels.ErrorResponse
// @Failure 404 {object} apiModels.ErrorResponse
// @Failure 500 {object} apiModels.ErrorResponse
// @Router /subscriptions/{id} [get]
func (ctrl *SubscriptionController) GetSubscriptionByID(ctx *gin.Context) {
	var id apiModels.ItemByIDRequest
	if err := ctx.ShouldBindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: apiModels.ErrBadParam.Error()})
		return
	}

	sub, err := ctrl.subscriptionService.GetSubscriptionByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, apiModels.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: service.ErrIES.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, sub)
}

// UpdateSubscriptionByID godoc
// @Summary Update a subscription
// @Description Updates an existing subscription record. Supports partial updates.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription UUID"
// @Param request body apiModels.UpdateSubscriptionRequest true "Subscription update data"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} apiModels.ErrorResponse
// @Failure 404 {object} apiModels.ErrorResponse
// @Failure 500 {object} apiModels.ErrorResponse
// @Router /subscriptions/{id} [put]
func (ctrl *SubscriptionController) UpdateSubscriptionByID(ctx *gin.Context) {
	var id apiModels.ItemByIDRequest
	if err := ctx.ShouldBindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: apiModels.ErrBadParam.Error()})
		return
	}

	var req apiModels.UpdateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: apiModels.ErrBadJSON.Error()})
		return
	}

	sub, err := ctrl.subscriptionService.UpdateSubscriptionByID(ctx, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, apiModels.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: service.ErrIES.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, sub)
}

// DeleteSubscriptionByID godoc
// @Summary Delete a subscription
// @Description Marks a subscription record as (soft-)deleted in the database
// @Tags subscriptions
// @Param id path string true "Subscription UUID"
// @Success 200 "OK"
// @Failure 400 {object} apiModels.ErrorResponse
// @Failure 404 {object} apiModels.ErrorResponse
// @Failure 500 {object} apiModels.ErrorResponse
// @Router /subscriptions/{id} [delete]
func (ctrl *SubscriptionController) DeleteSubscriptionByID(ctx *gin.Context) {
	var id apiModels.ItemByIDRequest
	if err := ctx.ShouldBindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: apiModels.ErrBadParam.Error()})
		return
	}

	err := ctrl.subscriptionService.DeleteSubscriptionByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, apiModels.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: service.ErrIES.Error()})
		}
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Description Returns a list of subscriptions with optional filtering by user ID and service name
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Service Name"
// @Success 200 {object} []models.Subscription
// @Failure 400 {object} apiModels.ErrorResponse
// @Failure 500 {object} apiModels.ErrorResponse
// @Router /subscriptions [get]
func (ctrl *SubscriptionController) ListSubscriptions(ctx *gin.Context) {
	var req apiModels.ListSubscriptionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		return
	}

	subs, err := ctrl.subscriptionService.ListSubscriptions(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrValidationError) {
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: service.ErrIES.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, subs)
}

// TotalSubscriptionsCost godoc
// @Summary Get total cost
// @Description Calculates total cost of subscriptions for a period
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Service Name"
// @Param start_date query string true "Start Date (MM-YYYY)"
// @Param end_date query string true "End Date (MM-YYYY)"
// @Success 200 {object} apiModels.TotalCostResponse
// @Failure 400 {object} apiModels.ErrorResponse
// @Failure 500 {object} apiModels.ErrorResponse
// @Router /subscriptions/total [get]
func (ctrl *SubscriptionController) TotalSubscriptionsCost(ctx *gin.Context) {
	var req apiModels.TotalCostRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := ctrl.subscriptionService.TotalSubscriptionsCost(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrValidationError) {
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: service.ErrIES.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
