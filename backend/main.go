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
	"backend/internal/email"
	"backend/internal/gateway"
	"backend/internal/historical_data"
	"backend/internal/infra/crypto"
	"backend/internal/infra/database/cloud_db"
	httpMiddlewares "backend/internal/infra/transport/http/middlewares"
	"backend/internal/real_time_data"
	"backend/internal/sensor"
	"backend/internal/shared/config"
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
	
	authzMiddleware *httpMiddlewares.AuthzMiddleware,

	gatewayController *gateway.GatewayController,
	userController *user.Controller,
	authController *auth.Controller,
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

	private := router.Group("/api/v1")
	private.Use(authzMiddleware.RequireAuthToken)

	// Auth
	{
		// Session
		public.POST("/auth/login", authController.LoginUser)
		private.POST("/auth/logout", authController.LogoutUser)

		// Conferma account
		public.GET("/auth/confirm_account/verify_token/{token}", authController.VerifyConfirmAccountToken)
		public.POST("/auth/confirm_account", authController.ConfirmAccount)

		// Password dimenticata
		public.GET("/auth/forgot_password/verify_token/{token}", authController.VerifyForgotPasswordToken)
		public.POST("/auth/forgot_password/request", authController.RequestForgotPasswordToken)
		public.POST("/auth/forgot_password", authController.ConfirmForgotPasswordToken)

		// Cambia password
		private.POST("/auth/change_password", authController.ChangePassword)
	}

	// User
	{
		private.POST("/tenant/:tenant_id/tenant_user", userController.CreateTenantUser)
		private.POST("/tenant/:tenant_id/tenant_admin", userController.CreateTenantAdmin)
		private.POST("/super_admin", userController.CreateSuperAdmin)

		private.DELETE("/tenant/:tenant_id/tenant_user/:user_id", userController.DeleteTenantUser)
		private.DELETE("/tenant/:tenant_id/tenant_admin/:user_id", userController.DeleteTenantAdmin)
		private.DELETE("/super_admin/:user_id", userController.DeleteSuperAdmin)

		private.GET("/tenant/:tenant_id/tenant_user/:user_id", userController.GetTenantUser)
		private.GET("/tenant/:tenant_id/tenant_admin/:user_id", userController.GetTenantAdmin)
		private.GET("/super_admin/:user_id", userController.GetSuperAdmin)

		private.GET("/tenant/:tenant_id/tenant_users", userController.GetTenantUsers)
		private.GET("/tenant/:tenant_id/tenant_admins", userController.GetTenantAdmins)
		private.GET("/super_admins", userController.GetSuperAdmins)
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
		crypto.Module,
		cloud_db.Module, // NOTA: Questo esegue la migrazione PRIMA di eseguire NewGinEngine()
		email.Module,
		httpMiddlewares.Module,

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
