package main

<<<<<<< HEAD

func main() {
	
}
=======
import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"

	// "backend/internal/common"
	"backend/internal/alert"
	"backend/internal/api_key"
	"backend/internal/auth"
	"backend/internal/common/db_connection"
	"backend/internal/common/migrate"
	"backend/internal/config"
	"backend/internal/email"
	"backend/internal/gateway"
	"backend/internal/historical_data"
	"backend/internal/real_time_data"
	"backend/internal/sensor"
	"backend/internal/tenant"
	"backend/internal/user"
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
	config *config.Config,
	gatewayController *gateway.GatewayController,
	userController *user.Controller,
) *gin.Engine {
	router := gin.Default()

	log.Info("CONFIG DB URL:" + config.CloudDBUrl)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Starting HTTP server!")

			public := router.Group("/api/v1")

			{
				public.GET("/", func(ctx *gin.Context) {
					ctx.JSON(200, gin.H{
						"msg": "It works!",
					})
				})

				// TODO: questo è solo un test
				public.POST("/gateway", gatewayController.CreateGateway)
				public.DELETE("/gateway/:id", gatewayController.DeleteGateway)

				public.POST("/tenant_user", userController.CreateTenantUser)
				public.POST("/tenant_admin", userController.CreateTenantAdmin)
				public.POST("/super_admin", userController.CreateSuperAdmin)

				public.DELETE("/tenant/:tenant_id/tenant_user/:user_id", userController.DeleteTenantUser)
				public.DELETE("/tenant/:tenant_id/tenant_admin/:user_id", userController.DeleteTenantAdmin)
				public.DELETE("/super_admin/:user_id", userController.DeleteSuperAdmin)

				public.GET("/tenant/:tenant_id/tenant_user/:user_id", userController.GetTenantUser)
				public.GET("/tenant/:tenant_id/tenant_admin/:user_id", userController.GetTenantAdmin)
				public.GET("/super_admin/:user_id", userController.GetSuperAdmin)
				public.POST("/users", userController.GetUsers)
				public.GET("/tenant/:tenant_id/users", userController.GetUsersByTenantId)
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
		// Moduli infrastrutturali
		config.Module,
		db_connection.Module,
		migrate.Module,
		email.Module,

		// Moduli funzionalità
		alert.Module,   // NOTA: Desiderabile
		api_key.Module, // NOTA: Desiderabile
		auth.Module,
		gateway.Module,
		historical_data.Module,
		real_time_data.Module,
		sensor.Module,
		tenant.Module,
		user.Module,

		// TODO: Spostare funzioni in rispettivi moduli???
		fx.Provide(
			NewGinEngine,
			zap.NewExample,
		),
		fx.Invoke(func(*gin.Engine) {}),
	).Run()
}
>>>>>>> origin/issue-14
