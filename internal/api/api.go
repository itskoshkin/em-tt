package api

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"subscription-aggregator-service/docs"
	ctrl "subscription-aggregator-service/internal/api/controllers"
	"subscription-aggregator-service/internal/api/middlewares"
	"subscription-aggregator-service/internal/config"
	"subscription-aggregator-service/internal/logger"
	"subscription-aggregator-service/internal/utils/graceful"
)

type API struct {
	engine *gin.Engine
	ctrl   *ctrl.SubscriptionController
}

func NewAPI(ctrl *ctrl.SubscriptionController) *API {
	if viper.GetBool(config.GinReleaseMode) && viper.GetString(config.LogLevel) != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.New()
	_ = e.SetTrustedProxies(nil) // Can nil produce an error? Or can a robot write a symphony?
	e.Use(gin.Recovery())
	e.Use(logger.GinLoggerMiddleware())
	e.Use(middlewares.RequestID())
	a := &API{engine: e, ctrl: ctrl}
	a.registerRoutes()
	return a
}

func (a *API) registerRoutes() {
	// API
	base := a.engine.Group(viper.GetString(config.ApiBasePath))
	{
		//subscriptions := base.Group("/subscriptions")
		{
			base.POST("/subscriptions", a.ctrl.CreateSubscription)
			base.GET("/subscriptions/total", a.ctrl.TotalSubscriptionsCost) // Must be above parameterized route to avoid conflict
			base.GET("/subscriptions/:id", a.ctrl.GetSubscriptionByID)
			base.PUT("/subscriptions/:id", a.ctrl.UpdateSubscriptionByID)
			base.DELETE("/subscriptions/:id", a.ctrl.DeleteSubscriptionByID)
			base.GET("/subscriptions", a.ctrl.ListSubscriptions)
		}
	}
	// Swagger
	{
		{
			docs.SwaggerInfo.Title = "Subscription Aggregator Service"
			docs.SwaggerInfo.Description = "CRUD API for managing user subscriptions"
			docs.SwaggerInfo.Version = "1.0"
			docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", viper.GetString(config.ApiHost), viper.GetString(config.ApiPort))
			docs.SwaggerInfo.BasePath = viper.GetString(config.ApiBasePath)
		}
		a.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func (a *API) Run() {
	address := fmt.Sprintf("%s:%s", viper.GetString(config.ApiHost), viper.GetString(config.ApiPort))
	fmt.Printf("API server listening on %s... \n", address)

	if err := graceful.RunGin(a.engine, address, viper.GetDuration(config.ApiShutdownTimeout)); err != nil {
		log.Printf("API server shutdown error: %v", err)
	}
}
