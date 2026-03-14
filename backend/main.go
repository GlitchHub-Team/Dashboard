package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"

	// "backend/internal/common"
	"backend/internal/alert"
	"backend/internal/api_key"
	"backend/internal/auth"
	"backend/internal/gateway"
	"backend/internal/historical_data"
	"backend/internal/real_time_data"
	"backend/internal/sensor"
	"backend/internal/tenant"
	"backend/internal/user"
	"context"
	"fmt"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error loading .env")
	}
}


func NewGinEngine(
	lc fx.Lifecycle,
	log *zap.Logger,
	gatewayController *gateway.GatewayController,
	userController *user.Controller,
) *gin.Engine {
	router := gin.Default()

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Starting HTTP server!")
			
			public := router.Group("/api")

			{
				// TODO: questo è solo un test
				public.POST("/gateway", gatewayController.CreateGateway)
				public.DELETE("/gateway", gatewayController.DeleteGateway)

				public.POST("/tenant_user", userController.CreateTenantUser)
				public.POST("/tenant_admin", userController.CreateTenantAdmin)
				public.POST("/super_admin", userController.CreateSuperAdmin)

			}

			go router.Run()
			return nil
		},
		OnStop: func(context.Context) error {
			log.Info("Stopping HTTP server!")
			return nil
		},
	})

	return router
}



func main() {

	fx.New(
		alert.Module,
		api_key.Module,
		auth.Module,
		gateway.Module,
		historical_data.Module,
		real_time_data.Module,
		sensor.Module,
		tenant.Module,
		user.Module,

		fx.Provide(
			NewGinEngine,
			zap.NewExample,
		),
		fx.Invoke(func(*gin.Engine) {}),
		
	).Run()
}
