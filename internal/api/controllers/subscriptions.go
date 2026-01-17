package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	apiModels "subscription-aggregator-service/internal/api/models"
	"subscription-aggregator-service/internal/service"
)

type SubscriptionController struct {
	subscriptionService service.SubscriptionService
}

func NewSubscriptionController(ss service.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{subscriptionService: ss}
}

func (ctrl *SubscriptionController) CreateSubscription(ctx *gin.Context) {
	var req apiModels.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()}) //TODO here and below: Hide internal error details? Or nothing sensitive exposed here?
		return
	}

	sub, err := ctrl.subscriptionService.CreateSubscription(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, sub)
}

func (ctrl *SubscriptionController) GetSubscriptionByID(ctx *gin.Context) {
	var id apiModels.ItemByIDRequest
	if err := ctx.ShouldBindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		return
	}

	sub, err := ctrl.subscriptionService.GetSubscriptionByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, apiModels.ErrorResponse{Error: "subscription not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, sub)
}

func (ctrl *SubscriptionController) UpdateSubscriptionByID(ctx *gin.Context) {
	var id apiModels.ItemByIDRequest
	if err := ctx.ShouldBindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		return
	}

	var req apiModels.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		return
	}

	sub, err := ctrl.subscriptionService.UpdateSubscriptionByID(ctx, id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, apiModels.ErrorResponse{Error: "subscription not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, sub)
}

func (ctrl *SubscriptionController) DeleteSubscriptionByID(ctx *gin.Context) {
	var id apiModels.ItemByIDRequest
	if err := ctx.ShouldBindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		return
	}

	err := ctrl.subscriptionService.DeleteSubscriptionByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationError):
			ctx.JSON(http.StatusBadRequest, apiModels.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrNotFound):
			ctx.JSON(http.StatusNotFound, apiModels.ErrorResponse{Error: "subscription not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, apiModels.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (ctrl *SubscriptionController) ListSubscriptions(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (ctrl *SubscriptionController) TotalSubscriptionsCost(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}