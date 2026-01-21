package app

import (
	"subscription-aggregator-service/internal/api"
	"subscription-aggregator-service/internal/api/controllers"
	"subscription-aggregator-service/internal/config"
	"subscription-aggregator-service/internal/logger"
	"subscription-aggregator-service/internal/service"
	"subscription-aggregator-service/internal/storage"
	"subscription-aggregator-service/pkg/postgres"
)

type App struct {
	API *api.API
}

func Load() *App {
	config.LoadConfig()
	logger.SetupLogger()
	db := postgres.NewInstance(config.DatabaseConfig())
	st := storage.NewSubscriptionsStorage(db)
	svc := service.NewSubscriptionService(st)
	ctrl := controllers.NewSubscriptionController(svc)
	return &App{API: api.NewAPI(ctrl)}
}

func (a *App) Run() {
	a.API.Run()
}
