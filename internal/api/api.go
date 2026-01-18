package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	ctrl "subscription-aggregator-service/internal/api/controllers"
	"subscription-aggregator-service/internal/config"
)

type API struct {
	engine *gin.Engine
	ctrl   *ctrl.SubscriptionController
}

func NewAPI(ctrl *ctrl.SubscriptionController) *API {
	if viper.GetBool(config.GinReleaseMode) {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.Default()
	a := &API{engine: e, ctrl: ctrl}
	a.registerRoutes()
	return a
}

func (a *API) registerRoutes() {
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
}

func (a *API) Run() {
	//TODO: Graceful shutdown?
	address := fmt.Sprintf("%s:%s", viper.GetString(config.ApiHost), viper.GetString(config.ApiPort))
	fmt.Printf("API server listening on %s... \n", address)
	if err := a.engine.Run(address); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Fatal: %s", err)
		}
	}
}
