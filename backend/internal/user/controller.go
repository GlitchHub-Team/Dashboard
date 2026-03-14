package user

import (
	"net/http"

	"backend/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	createTenantUserUseCase   CreateTenantUserUseCase
	createTenantAdminUseCase  CreateTenantAdminUseCase
	createSuperAdminUseCase   CreateSuperAdminUseCase
	deleteUserUseCase         DeleteUserUseCase
	getUserByIdUseCase        GetUserByIdUseCase
	getUsersByTenantIdUseCase GetUsersByTenantIdUseCase
	getUsersUseCase           GetUsersUseCase
}

func NewUserController(
	createTenantUserUseCase CreateTenantUserUseCase,
	createTenantAdminUseCase CreateTenantAdminUseCase,
	createSuperAdminUseCase CreateSuperAdminUseCase,
	deleteUserUseCase DeleteUserUseCase,
	getUserByIdUseCase GetUserByIdUseCase,
	getUsersByTenantIdUseCase GetUsersByTenantIdUseCase,
	getUsersUseCase GetUsersUseCase,
) *Controller {
	return &Controller{
		createTenantUserUseCase:   createTenantUserUseCase,
		createTenantAdminUseCase:  createTenantAdminUseCase,
		createSuperAdminUseCase:   createSuperAdminUseCase,
		deleteUserUseCase:         deleteUserUseCase,
		getUserByIdUseCase:        getUserByIdUseCase,
		getUsersByTenantIdUseCase: getUsersByTenantIdUseCase,
		getUsersUseCase:           getUsersUseCase,
	}
}

func (controller *Controller) CreateTenantUser(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto CreateTenantUserDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := CreateTenantUserCommand{
		Email:    requestDto.Email,
		Username: requestDto.Username,
		TenantId: tenantId,
	}

	user, err := controller.createTenantUserUseCase.CreateTenantUser(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) CreateTenantAdmin(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto CreateTenantAdminDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := CreateTenantAdminCommand{
		Email:    requestDto.Email,
		Username: requestDto.Username,
		TenantId: tenantId,
	}

	user, err := controller.createTenantAdminUseCase.CreateTenantAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) CreateSuperAdmin(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto CreateSuperAdminDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := CreateSuperAdminCommand{
		Email:    requestDto.Email,
		Username: requestDto.Username,
	}

	user, err := controller.createSuperAdminUseCase.CreateSuperAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteUser(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto DeleteUserDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := DeleteUserCommand{
		UserId: requestDto.UserId,
	}

	oldUser, err := controller.deleteUserUseCase.DeleteUser(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetUserById(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto GetUserByIdDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := GetUserByIdCommand{
		UserId: requestDto.UserId,
	}

	user, err := controller.getUserByIdUseCase.GetUserById(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetUsers(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto GetUsersDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := GetUsersCommand{
		Page:  requestDto.Page,
		Limit: requestDto.Limit,
		Role:  (UserRole)(requestDto.UserRole),
	}

	users, total, err := controller.getUsersUseCase.GetUsers(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetUsersByTenantId(ctx *gin.Context) {
}
