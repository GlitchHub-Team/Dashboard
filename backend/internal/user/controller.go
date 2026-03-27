package user

import (
	"fmt"
	"net/http"

	"backend/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Controller struct {
	log *zap.Logger

	createTenantUserUseCase   CreateTenantUserUseCase
	createTenantAdminUseCase  CreateTenantAdminUseCase
	createSuperAdminUseCase   CreateSuperAdminUseCase
	deleteTenantUserUseCase   DeleteTenantUserUseCase
	deleteTenantAdminUseCase  DeleteTenantAdminUseCase
	deleteSuperAdminUseCase   DeleteSuperAdminUseCase
	getTenantUserUseCase      GetTenantUserUseCase
	getTenantAdminUseCase     GetTenantAdminUseCase
	getSuperAdminUseCase      GetSuperAdminUseCase
	getUsersByTenantIdUseCase GetUsersByTenantIdUseCase
	getUsersUseCase           GetUsersUseCase
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
	getUsersByTenantIdUseCase GetUsersByTenantIdUseCase,
	getUsersUseCase GetUsersUseCase,
) *Controller {
	return &Controller{
		log: log,

		createTenantUserUseCase:  createTenantUserUseCase,
		createTenantAdminUseCase: createTenantAdminUseCase,
		createSuperAdminUseCase:  createSuperAdminUseCase,

		deleteTenantUserUseCase:  deleteTenantUserUseCase,
		deleteTenantAdminUseCase: deleteTenantAdminCase,
		deleteSuperAdminUseCase:  deleteSuperAdminCase,

		getTenantUserUseCase:      getTenantUserUseCase,
		getTenantAdminUseCase:     getTenantAdminUseCase,
		getSuperAdminUseCase:      getSuperAdminUseCase,
		getUsersByTenantIdUseCase: getUsersByTenantIdUseCase,
		getUsersUseCase:           getUsersUseCase,
	}
}

// Create =============================================================================
func (controller *Controller) CreateTenantUser(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto CreateTenantUserDTO

	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, fmt.Errorf("error parsing tenant_id: %v", err))
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
		common.RequestError(ctx, fmt.Errorf("error creating tenant user: %v", err))
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

// func (controller *Controller) DeleteUser(ctx *gin.Context) {
// 	// Binding e validazione input
// 	var requestDto DeleteUserDTO

// 	// TODO: Leggere parametri da input string invece che da body POST
// 	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
// 		common.RequestError(ctx, err)
// 		return
// 	}

// 	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

// 	// Esecuzione comando
// 	cmd := DeleteUserCommand{
// 		UserId: requestDto.UserId,
// 	}

// 	oldUser, err := controller.deleteUserUseCase.DeleteUser(cmd)
// 	if err != nil {
// 		common.RequestError(ctx, err)
// 		return
// 	}

// 	// Risposta
// 	responseDto := NewUserResponseDTO(oldUser)
// 	ctx.JSON(http.StatusOK, responseDto)
// }

func (controller *Controller) DeleteTenantUser(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto DeleteTenantUserDTO

	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Esecuzione comando
	cmd := DeleteTenantUserCommand{
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	oldUser, err := controller.deleteTenantUserUseCase.DeleteTenantUser(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteTenantAdmin(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto DeleteTenantAdminDTO

	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Esecuzione comando
	cmd := DeleteTenantAdminCommand{
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	oldUser, err := controller.deleteTenantAdminUseCase.DeleteTenantAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteSuperAdmin(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto DeleteSuperAdminDTO

	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := DeleteSuperAdminCommand{
		UserId: requestDto.UserId,
	}

	oldUser, err := controller.deleteSuperAdminUseCase.DeleteSuperAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

// func (controller *Controller) GetUserById(ctx *gin.Context) {
// 	// Binding e validazione input
// 	var requestDto GetUserByIdDTO

// 	// TODO: Leggere parametri da input string invece che da body POST
// 	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
// 		common.RequestError(ctx, err)
// 		return
// 	}

// 	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

// 	// Esecuzione comando
// 	cmd := GetUserByIdCommand{
// 		UserId: requestDto.UserId,
// 	}

// 	user, err := controller.getUserByIdUseCase.GetUserById(cmd)
// 	if err != nil {
// 		common.RequestError(ctx, err)
// 		return
// 	}

// 	// Risposta
// 	responseDto := NewUserResponseDTO(user)
// 	ctx.JSON(http.StatusOK, responseDto)
// }

func (controller *Controller) GetTenantUser(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto GetTenantUserDTO

	if err := ctx.ShouldBindUri(&requestDto); err != nil {
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
	cmd := GetTenantUserCommand{
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	user, err := controller.getTenantUserUseCase.GetTenantUser(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: remove debug
	if user.IsZero() {
		// common.RequestError(ctx, errUserCreationFailed)
		ctx.JSON(200, gin.H{
			"tenant_id": cmd.TenantId,
			"user_id":   cmd.UserId,
		})
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetTenantAdmin(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto GetTenantAdminDTO

	// TODO: Leggere parametri da input string invece che da body POST
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
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
	cmd := GetTenantAdminCommand{
		TenantId: tenantId,
		UserId:   requestDto.UserId,
	}

	user, err := controller.getTenantAdminUseCase.GetTenantAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	common.RequestOk(ctx, responseDto)
}

func (controller *Controller) GetSuperAdmin(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto GetSuperAdminDTO

	// TODO: Leggere parametri da input string invece che da body POST
	if err := ctx.ShouldBindUri(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := GetSuperAdminCommand{
		UserId: requestDto.UserId,
	}

	user, err := controller.getSuperAdminUseCase.GetSuperAdmin(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserResponseDTO(user)
	common.RequestOk(ctx, responseDto)
}

func (controller *Controller) GetUsers(ctx *gin.Context) {
	// Binding e validazione input
	var requestDto GetUsersDTO

	// TODO: Leggere parametri da input string invece che da body POST
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
	// Binding e validazione input
	var requestDto GetUsersByTenantIdDTO

	// TODO: Leggere parametri da input string invece che da body POST
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		common.RequestError(ctx, err)
		return
	}

	tenantId, err := uuid.Parse(requestDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
	}

	// TODO: verificare che l'utente abbia autorizzazione per fare richiesta!

	// Esecuzione comando
	cmd := GetUsersByTenantIdCommand{
		Page:     requestDto.Page,
		Limit:    requestDto.Limit,
		TenantId: tenantId,
	}

	users, total, err := controller.getUsersByTenantIdUseCase.GetUsersByTenantId(cmd)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	// Risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}
