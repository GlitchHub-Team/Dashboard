package gateway

import (
	// "net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)


type GatewayController struct {
	log *zap.Logger

	createGatewayUseCase CreateGatewayUseCase
	deleteGatewayUseCase DeleteGatewayUseCase
	// ...
}

func NewGatewayController(
	log *zap.Logger,
	createGatewayUseCase CreateGatewayUseCase,
	deleteGatewayUseCase DeleteGatewayUseCase,
) *GatewayController {
	return &GatewayController{
		log,
		createGatewayUseCase,
		deleteGatewayUseCase,
	}
}

func (c *GatewayController) CreateGateway(ctx *gin.Context, ) {

	// var dto createGatewayDTO
	// if err := ctx.ShouldBindJSON(&dto); err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{
	// 		"error": err.Error(),
	// 	})
	// 	return
	// }

	// cmd := CreateGatewayCommand{
	// 	Name: dto.Name,
	// }
	// gateway, err := c.createGatewayUseCase.CreateGateway(cmd)
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{
	// 		"error": err.Error(),
	// 	})
	// 	return
	// }

	// response := gatewayResponseDTO{
	// 	Id:		gateway.Id.String(),
	// 	Name: 	gateway.Name,
	// 	Status: string(gateway.Status),
	// }
	// ctx.JSON(http.StatusOK, response)
}

func (c *GatewayController) DeleteGateway(ctx *gin.Context, ) {
	// var command deleteGatewayDTO
}


