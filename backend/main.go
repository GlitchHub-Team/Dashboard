package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

	// https://blog.depa.do/post/gin-validation-errors-handling
	if validator, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	} else {
		log.Warn("Cannot register name func!")
	}

	public := router.Group("/api/v1")

	{
		// TODO: questo è solo un test
		public.POST("/gateway", gatewayController.CreateGateway)
		public.DELETE("/gateway/:id", gatewayController.DeleteGateway)

		public.POST("/tenant/:tenant_id/tenant_user", userController.CreateTenantUser)
		public.POST("/tenant/:tenant_id/tenant_admin", userController.CreateTenantAdmin)
		public.POST("/super_admin", userController.CreateSuperAdmin)

		public.DELETE("/tenant/:tenant_id/tenant_user/:user_id", userController.DeleteTenantUser)
		public.DELETE("/tenant/:tenant_id/tenant_admin/:user_id", userController.DeleteTenantAdmin)
		public.DELETE("/super_admin/:user_id", userController.DeleteSuperAdmin)

		public.GET("/tenant/:tenant_id/tenant_user/:user_id", userController.GetTenantUser)
		public.GET("/tenant/:tenant_id/tenant_admin/:user_id", userController.GetTenantAdmin)
		public.GET("/super_admin/:user_id", userController.GetSuperAdmin)

		public.GET("/tenant/:tenant_id/tenant_users", userController.GetTenantUsers)
		public.GET("/tenant/:tenant_id/tenant_admins", userController.GetTenantAdmins)
		public.GET("/super_admins", userController.GetSuperAdmins)

	}

	log.Info("CONFIG DB URL:" + config.CloudDBUrl)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Starting HTTP server!")
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
		migrate.Module, // NOTA: Questo esegue la migrazione PRIMA di eseguire NewGinEngine()
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
