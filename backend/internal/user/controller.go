package user

import (
	"fmt"
	"net/http"

	"backend/internal/common"
	"backend/internal/common/dto"
	// "backend/internal/identity"
	transportHttp "backend/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Controller struct {
	log *zap.Logger

	createTenantUserUseCase  CreateTenantUserUseCase
	createTenantAdminUseCase CreateTenantAdminUseCase
	createSuperAdminUseCase  CreateSuperAdminUseCase

	deleteTenantUserUseCase  DeleteTenantUserUseCase
	deleteTenantAdminUseCase DeleteTenantAdminUseCase
	deleteSuperAdminUseCase  DeleteSuperAdminUseCase

	getTenantUserUseCase  GetTenantUserUseCase
	getTenantAdminUseCase GetTenantAdminUseCase
	getSuperAdminUseCase  GetSuperAdminUseCase

	getTenantUsersByTenantUseCase  GetTenantUsersByTenantUseCase
	getTenantAdminsByTenantUseCase GetTenantAdminsByTenantUseCase
	getSuperAdminListUseCase       GetSuperAdminListUseCase
	// getUsersUseCase           GetUsersUseCase
}

func NewUserController(
	log *zap.Logger,
	createTenantUserUseCase CreateTenantUserUseCase,
	createTenantAdminUseCase CreateTenantAdminUseCase,
	createSuperAdminUseCase CreateSuperAdminUseCase,

	deleteTenantUserUseCase DeleteTenantUserUseCase,
	deleteTenantAdminCase DeleteTenantAdminUseCase,
	deleteSuperAdminCase DeleteSuperAdminUseCase,

	getTenantUserUseCase GetTenantUserUseCase,
	getTenantAdminUseCase GetTenantAdminUseCase,
	getSuperAdminUseCase GetSuperAdminUseCase,

	getTenantUsersByTenantUseCase GetTenantUsersByTenantUseCase,
	getTenantAdminsByTenantUseCase GetTenantAdminsByTenantUseCase,
	getSuperAdminListUseCase GetSuperAdminListUseCase,
	// getUsersUseCase GetUsersUseCase,
) *Controller {
	return &Controller{
		log: log,

		createTenantUserUseCase:  createTenantUserUseCase,
		createTenantAdminUseCase: createTenantAdminUseCase,
		createSuperAdminUseCase:  createSuperAdminUseCase,

		deleteTenantUserUseCase:  deleteTenantUserUseCase,
		deleteTenantAdminUseCase: deleteTenantAdminCase,
		deleteSuperAdminUseCase:  deleteSuperAdminCase,

		getTenantUserUseCase:  getTenantUserUseCase,
		getTenantAdminUseCase: getTenantAdminUseCase,
		getSuperAdminUseCase:  getSuperAdminUseCase,

		getTenantUsersByTenantUseCase:  getTenantUsersByTenantUseCase,
		getTenantAdminsByTenantUseCase: getTenantAdminsByTenantUseCase,
		getSuperAdminListUseCase:       getSuperAdminListUseCase,
		// getUsersUseCase:           getUsersUseCase,
	}
}

// Create =============================================================================================

func (controller *Controller) CreateTenantUser(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto CreateTenantUserDTO
	
	// 2. Binding JSON
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, _ := uuid.Parse(requestDto.TenantId)

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 3. Esecuzione comando
	cmd := CreateTenantUserCommand{
		Requester: requester,
		Email:    requestDto.Email,
		Username: requestDto.Username,
		TenantId: tenantId,
	}

	user, err := controller.createTenantUserUseCase.CreateTenantUser(cmd)
	if err != nil {
		common.RequestError(ctx, fmt.Errorf("error creating tenant user: %v", err))
		return
	}

	// 4. Invio risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) CreateTenantAdmin(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto CreateTenantAdminDTO
	
	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 3. Binding JSON
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, _ := uuid.Parse(requestDto.TenantId)

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 4. Esecuzione comando
	cmd := CreateTenantAdminCommand{
		Requester: requester,
		Email:    requestDto.Email,
		Username: requestDto.Username,
		TenantId: tenantId,
	}

	user, err := controller.createTenantAdminUseCase.CreateTenantAdmin(cmd)
	if err != nil { 
		common.RequestError(ctx, err)
		return
	}

	// 5. Invio risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) CreateSuperAdmin(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}
	
	var requestDto CreateSuperAdminDTO
	
	// 2. Binding JSON
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 3. Esecuzione comando
	cmd := CreateSuperAdminCommand{
		Requester: requester,
		Email:    requestDto.Email,
		Username: requestDto.Username,
	}

	user, err := controller.createSuperAdminUseCase.CreateSuperAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 4. Invio risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

// Delete =============================================================================================

func (controller *Controller) DeleteTenantUser(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto DeleteTenantUserDTO
	
	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 3. Binding JSON
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	tenantId, _ := uuid.Parse(requestDto.TenantId)


	// 4. Esecuzione comando
	cmd := DeleteTenantUserCommand{
		Requester: requester,
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	oldUser, err := controller.deleteTenantUserUseCase.DeleteTenantUser(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 5. Invio risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteTenantAdmin(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto DeleteTenantAdminDTO

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 3. Binding JSON
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 4. Esecuzione comando
	cmd := DeleteTenantAdminCommand{
		Requester: requester,
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	oldUser, err := controller.deleteTenantAdminUseCase.DeleteTenantAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 5. Invio risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteSuperAdmin(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}
	
	var requestDto DeleteSuperAdminDTO

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 3. Esecuzione comando
	cmd := DeleteSuperAdminCommand{
		Requester: requester,
		UserId: requestDto.UserId,
	}

	oldUser, err := controller.deleteSuperAdminUseCase.DeleteSuperAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 4. Invio risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

// Get single =========================================================================================

func (controller *Controller) GetTenantUser(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto GetTenantUserDTO

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, _ := uuid.Parse(requestDto.TenantId)

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 3. Esecuzione comando
	cmd := GetTenantUserCommand{
		Requester: requester,
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	user, err := controller.getTenantUserUseCase.GetTenantUser(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 4. Invio risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetTenantAdmin(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto GetTenantAdminDTO

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, _ := uuid.Parse(requestDto.TenantId)

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 3. Esecuzione comando
	cmd := GetTenantAdminCommand{
		Requester: requester,
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	user, err := controller.getTenantAdminUseCase.GetTenantAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 4. Invio risposta
	responseDto := NewUserResponseDTO(user)
	common.RequestOk(ctx, responseDto)
}

func (controller *Controller) GetSuperAdmin(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto GetSuperAdminDTO

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// 3. Esecuzione comando
	cmd := GetSuperAdminCommand{
		Requester: requester,
		UserId: requestDto.UserId,
	}

	user, err := controller.getSuperAdminUseCase.GetSuperAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 4. Invio risposta
	responseDto := NewUserResponseDTO(user)
	common.RequestOk(ctx, responseDto)
}


// Get multiple =======================================================================================

func (controller *Controller) GetTenantUsers(ctx *gin.Context) {	
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	requestDto := GetTenantUsersByTenantDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 3. Binding Query
	if err := ctx.ShouldBindQuery(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta! Da fare nel servizio!!!

	// 4. Esecuzione comando
	tenantId, _ := uuid.Parse(requestDto.TenantId)

	cmd := GetTenantUsersByTenantCommand{
		Requester: requester,
		Page:  requestDto.Page,
		Limit: requestDto.Limit,
		TenantId: tenantId,
	}

	users, total, err := controller.getTenantUsersByTenantUseCase.GetTenantUsersByTenant(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 5. Invio risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}


func (controller *Controller) GetTenantAdmins(ctx *gin.Context) {	
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	requestDto := GetTenantAdminsByTenantDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 3. Binding Query
	if err := ctx.ShouldBindQuery(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta! Da fare nel servizio!!!

	// 4. Esecuzione comando
	tenantId, _ := uuid.Parse(requestDto.TenantId)

	cmd := GetTenantAdminsByTenantCommand{
		Requester: requester,
		Page:  requestDto.Page,
		Limit: requestDto.Limit,
		TenantId: tenantId,
	}

	users, total, err := controller.getTenantAdminsByTenantUseCase.GetTenantAdminsByTenant(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 5. Invio risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}


func (controller *Controller) GetSuperAdmins(ctx *gin.Context) {
	// 1. Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	requestDto := GetTenantUsersByTenantDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}

	// 2. Binding URI
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 3. Binding Query
	if err := ctx.ShouldBindQuery(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta! Da fare nel servizio!!!

	// 4. Esecuzione comando

	cmd := GetSuperAdminListCommand{
		Requester: requester,
		Page:  requestDto.Page,
		Limit: requestDto.Limit,
	}

	users, total, err := controller.getSuperAdminListUseCase.GetSuperAdminList(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// 5. Invio risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}
