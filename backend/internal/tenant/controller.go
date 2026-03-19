package tenant

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller struct{
	log *zap.Logger
	createTenantUseCase CreateTenantUseCase
	deleteTenantUseCase DeleteTenantUseCase
	getTenantUseCase GetTenantUseCase
	getTenantListUseCase GetTenantListUseCase
	getTenantByUserUseCase GetTenantByUserUseCase
}

func NewTenantController(
	log *zap.Logger,
	createTenantUseCase CreateTenantUseCase,
	deleteTenantUseCase DeleteTenantUseCase,
	getTenantUseCase GetTenantUseCase,
	getTenantListUseCase GetTenantListUseCase,
	getTenantByUserUseCase GetTenantByUserUseCase,
	) *Controller {
	return &Controller{
		log: log,
		createTenantUseCase: createTenantUseCase,
		deleteTenantUseCase: deleteTenantUseCase,
		getTenantUseCase: getTenantUseCase,
		getTenantListUseCase: getTenantListUseCase,
		getTenantByUserUseCase: getTenantByUserUseCase,
	}
}

//CREATE TENANT =======================================================================================

func (controller *Controller) CreateTenant(ctx *gin.Context){
	var cmd CreateTenantCommand

	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	tenant, err := controller.createTenantUseCase.CreateTenant(cmd)

	if err != nil {
		controller.log.Error("Error creating tenant", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to create tenant"))
		return
	}

	ctx.JSON(http.StatusOK, tenant)
}

//DELETE TENANT =======================================================================================

func (controller *Controller) DeleteTenant(ctx *gin.Context){
	var cmd DeleteTenantCommand

	if err:= ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	oldTenant, err := controller.deleteTenantUseCase.DeleteTenant(cmd)

	if err != nil {
		controller.log.Error("Error deleting tenant", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to delete tenant"))
		return
	}

	ctx.JSON(http.StatusOK, oldTenant)
}

//GET TENANT ==========================================================================================

func (controller *Controller) GetTenant(ctx *gin.Context){
	var cmd GetTenantCommand

	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	tenant, err := controller.getTenantUseCase.GetTenant(cmd)

	if err != nil {
		controller.log.Error("Error getting tenant", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to get tenant"))
		return
	}

	ctx.JSON(http.StatusOK, tenant)
}

//GET TENANT ==========================================================================================

func (controller *Controller) GetTenants(ctx *gin.Context){
	var cmd GetTenantListCommand

	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	tenants, err := controller.getTenantListUseCase.GetTenantList(cmd)

	if err != nil {
		controller.log.Error("Error getting tenants", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to get tenants"))
		return
	}

	ctx.JSON(http.StatusOK, tenants)
}

//GET TENANT BY USER ==================================================================================

func (controller *Controller) GetTenantByUser(ctx *gin.Context){
	var cmd GetTenantByUserCommand

	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	tenant, err := controller.getTenantByUserUseCase.GetTenantByUser(cmd)

	if err != nil {
		controller.log.Error("Error getting tenant by user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to get tenant by user"))
		return
	}

	ctx.JSON(http.StatusOK, tenant)
}
