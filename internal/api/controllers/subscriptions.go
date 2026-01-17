package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"subscription-aggregator-service/internal/service"
)

type SubscriptionController struct {
	subscriptionService service.SubscriptionService
}

func NewSubscriptionController(ss service.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{subscriptionService: ss}
}

func (c *SubscriptionController) Create(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (c *SubscriptionController) Get(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (c *SubscriptionController) Update(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (c *SubscriptionController) Delete(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (c *SubscriptionController) List(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (c *SubscriptionController) TotalCost(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
