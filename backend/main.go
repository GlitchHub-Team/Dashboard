package main

import (
	"context"

	"backend/internal/infra/modules"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	// "backend/internal/common"
)

func Start(
	lc fx.Lifecycle,
	router *gin.Engine,
	log *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Starting HTTP server!")
			go router.Run() //nolint:errcheck
			return nil
		},
		OnStop: func(context.Context) error {
			log.Info("Stopping HTTP server!")
			return nil
		},
	})
}

func main() {
	fx.New(
		modules.AppModules(),
		fx.Invoke(Start),
	).Run()
}
