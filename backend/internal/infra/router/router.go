package router

import (
	"reflect"
	"strings"

	"backend/internal/auth"
	"backend/internal/gateway"
	"backend/internal/historical_data"
	"backend/internal/real_time_data"
	"backend/internal/sensor"
	"backend/internal/shared/config"
	"backend/internal/tenant"
	"backend/internal/user"

	httpMiddlewares "backend/internal/infra/transport/http/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func NewGinEngine(
	log *zap.Logger,
	config *config.Config,

	authzMiddleware *httpMiddlewares.AuthzMiddleware,

	gatewayController *gateway.GatewayController,
	historicalDataController *historical_data.Controller,
	realTimeDataController *real_time_data.Controller,
	userController *user.Controller,
	authController *auth.Controller,
	sensorController *sensor.Controller,
	tenantController *tenant.Controller,
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

	corsConfig := cors.Config{
		AllowOrigins:     []string{config.AppURL,},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	router.Use(cors.New(corsConfig))

	public := router.Group("/api/v1")

	private := router.Group("/api/v1")
	private.Use(authzMiddleware.RequireAuthToken)

	websocketPrivate := router.Group("/api/v1")
	websocketPrivate.Use(authzMiddleware.RequireAuthTokenInQuery)

	// Auth
	{
		// Session
		public.POST("/auth/login", authController.LoginUser)
		private.POST("/auth/logout", authController.LogoutUser)

		// Conferma account
		public.POST("/auth/confirm_account/verify_token", authController.VerifyConfirmAccountToken)
		public.POST("/auth/confirm_account", authController.ConfirmAccount)

		// Password dimenticata
		public.POST("/auth/forgot_password/verify_token", authController.VerifyForgotPasswordToken)
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

	// Sensor
	{
		private.POST("/sensor", sensorController.CreateSensor)
		private.DELETE("/sensor/:sensor_id", sensorController.DeleteSensor)

		private.GET("/sensor/:sensor_id", sensorController.GetSensor)
		private.GET("/gateway/:gateway_id/sensors", sensorController.GetSensorsByGateway)
		private.GET("/tenant/:tenant_id/sensors", sensorController.GetSensorsByTenant)

		private.POST("/sensor/:sensor_id/interrupt", sensorController.InterruptSensor)
		private.POST("/sensor/:sensor_id/resume", sensorController.ResumeSensor)
	}

	// Historical data
	{
		private.GET(
			"/tenant/:tenant_id/sensor/:sensor_id/historical_data",
			historicalDataController.GetSensorHistoricalData,
		)
	}

	// Real time data
	{
		websocketPrivate.GET(
			"/tenant/:tenant_id/sensor/:sensor_id/real_time_data",
			realTimeDataController.GetRealTimeData,
		)
	}

	// Tenant
	{
		private.POST("/tenant", tenantController.CreateTenant)
		private.DELETE("/tenant/:tenant_id", tenantController.DeleteTenant)
		private.GET("/tenant/:tenant_id", tenantController.GetTenant)
		private.GET("/tenants", tenantController.GetTenantList)
		public.GET("/all_tenants", tenantController.GetAllTenants)
	}

	// Gateway
	{
		private.POST("/gateway", gatewayController.CreateGateway)
		private.DELETE("/gateway/:gateway_id", gatewayController.DeleteGateway)
		private.GET("/gateway/:gateway_id", gatewayController.GetGateway)
		private.GET("/gateways", gatewayController.GetAllGateways)
		// private.GET("/tenant/:tenant_id/gateway/:gateway_id", gatewayController.GetGatewayByTenant)
		private.GET("/tenant/:tenant_id/gateways", gatewayController.GetGatewaysByTenant)

		// Commands
		private.POST("/gateway/:gateway_id/commission", gatewayController.CommissionGateway)
		private.POST("/gateway/:gateway_id/decommission", gatewayController.DecommissionGateway)
		private.POST("/gateway/:gateway_id/interrupt", gatewayController.InterruptGateway)
		private.POST("/gateway/:gateway_id/resume", gatewayController.ResumeGateway)
		private.POST("/gateway/:gateway_id/reset", gatewayController.ResetGateway)
		private.POST("/gateway/:gateway_id/reboot", gatewayController.RebootGateway)
	}
	return router
}
