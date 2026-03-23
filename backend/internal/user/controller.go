package user

import (
	"errors"
	"net/http"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/user/mocks/use_cases_create.go -package=mocks . CreateTenantUserUseCase,CreateTenantAdminUseCase,CreateSuperAdminUseCase
//go:generate mockgen -destination=../../tests/user/mocks/use_cases_delete.go -package=mocks . DeleteTenantUserUseCase,DeleteTenantAdminUseCase,DeleteSuperAdminUseCase
//go:generate mockgen -destination=../../tests/user/mocks/use_cases_get.go -package=mocks . GetTenantUserUseCase,GetTenantAdminUseCase,GetSuperAdminUseCase,GetTenantUsersByTenantUseCase,GetTenantAdminsByTenantUseCase,GetSuperAdminListUseCase


// Create ----------------------------------------------------------

type CreateTenantUserUseCase interface {
	CreateTenantUser(cmd CreateTenantUserCommand) (User, error)
}
type CreateTenantAdminUseCase interface {
	CreateTenantAdmin(cmd CreateTenantAdminCommand) (User, error)
}
type CreateSuperAdminUseCase interface {
	CreateSuperAdmin(cmd CreateSuperAdminCommand) (User, error)
}

// Delete ----------------------------------------------------------

type DeleteTenantUserUseCase interface {
	DeleteTenantUser(cmd DeleteTenantUserCommand) (User, error)
}
type DeleteTenantAdminUseCase interface {
	DeleteTenantAdmin(cmd DeleteTenantAdminCommand) (User, error)
}
type DeleteSuperAdminUseCase interface {
	DeleteSuperAdmin(cmd DeleteSuperAdminCommand) (User, error)
}

// Get singolo -----------------------------------------------------

type GetTenantUserUseCase interface {
	GetTenantUser(cmd GetTenantUserCommand) (User, error)
}
type GetTenantAdminUseCase interface {
	GetTenantAdmin(cmd GetTenantAdminCommand) (User, error)
}
type GetSuperAdminUseCase interface {
	GetSuperAdmin(cmd GetSuperAdminCommand) (User, error)
}

// Get multiplo -----------------------------------------------------

type GetTenantUsersByTenantUseCase interface {
	GetTenantUsersByTenant(cmd GetTenantUsersByTenantCommand) (
		tenantUsers []User, total uint, err error,
	)
}
type GetTenantAdminsByTenantUseCase interface {
	GetTenantAdminsByTenant(cmd GetTenantAdminsByTenantCommand) (
		tenantAdmins []User, total uint, err error,
	)
}
type GetSuperAdminListUseCase interface {
	GetSuperAdminList(cmd GetSuperAdminListCommand) (
		superAdmins []User, total uint, err error,
	)
}


/*
Controller per il CRUD sugli utenti
*/
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
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Binding JSON
	var bodyDto CreateUserBodyDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := CreateTenantUserCommand{
		Requester: requester,
		Email:     bodyDto.Email,
		Username:  bodyDto.Username,
		TenantId:  tenantId,
	}

	createdUser, err := controller.createTenantUserUseCase.CreateTenantUser(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserAlreadyExists) {
			transportHttp.RequestError(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(createdUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) CreateTenantAdmin(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Binding JSON
	var bodyDto CreateUserBodyDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := CreateTenantAdminCommand{
		Requester: requester,
		Email:     bodyDto.Email,
		Username:  bodyDto.Username,
		TenantId:  tenantId,
	}

	createdUser, err := controller.createTenantAdminUseCase.CreateTenantAdmin(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserAlreadyExists) {
			transportHttp.RequestError(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(createdUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) CreateSuperAdmin(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var requestDto CreateUserBodyDTO

	// Binding JSON
	if err := ctx.ShouldBindJSON(&requestDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := CreateSuperAdminCommand{
		Requester: requester,
		Email:     requestDto.Email,
		Username:  requestDto.Username,
	}

	createdUser, err := controller.createSuperAdminUseCase.CreateSuperAdmin(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserAlreadyExists) {
			transportHttp.RequestError(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(createdUser)
	ctx.JSON(http.StatusOK, responseDto)
}

// Delete =============================================================================================

func (controller *Controller) DeleteTenantUser(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantMemberUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Esecuzione comando
	cmd := DeleteTenantUserCommand{
		Requester: requester,
		TenantId:  tenantId,
		UserId:    uriDto.UserId,
	}

	oldUser, err := controller.deleteTenantUserUseCase.DeleteTenantUser(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteTenantAdmin(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantMemberUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Esecuzione comando
	cmd := DeleteTenantAdminCommand{
		Requester: requester,
		TenantId:  tenantId,
		UserId:    uriDto.UserId,
	}

	oldUser, err := controller.deleteTenantAdminUseCase.DeleteTenantAdmin(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) DeleteSuperAdmin(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.SuperAdminUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := DeleteSuperAdminCommand{
		Requester: requester,
		UserId:    uriDto.UserId,
	}

	oldUser, err := controller.deleteSuperAdminUseCase.DeleteSuperAdmin(cmd)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, ErrUserNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(oldUser)
	ctx.JSON(http.StatusOK, responseDto)
}

// Get single =========================================================================================

func (controller *Controller) GetTenantUser(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantMemberUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Esecuzione comando
	cmd := GetTenantUserCommand{
		Requester: requester,
		TenantId:  tenantId,
		UserId:    uriDto.UserId,
	}

	user, err := controller.getTenantUserUseCase.GetTenantUser(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(user)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetTenantAdmin(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantMemberUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Esecuzione comando
	cmd := GetTenantAdminCommand{
		Requester: requester,
		TenantId:  tenantId,
		UserId:    uriDto.UserId,
	}

	user, err := controller.getTenantAdminUseCase.GetTenantAdmin(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(user)
	transportHttp.RequestOk(ctx, responseDto)
}

func (controller *Controller) GetSuperAdmin(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.SuperAdminUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := GetSuperAdminCommand{
		Requester: requester,
		UserId:    uriDto.UserId,
	}

	user, err := controller.getSuperAdminUseCase.GetSuperAdmin(cmd)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, ErrUserNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserResponseDTO(user)
	transportHttp.RequestOk(ctx, responseDto)
}

// Get multiple =======================================================================================

func (controller *Controller) GetTenantUsers(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Binding Query
	queryDto := GetUserListQueryDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := GetTenantUsersByTenantCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
		TenantId:  tenantId,
	}

	users, total, err := controller.getTenantUsersByTenantUseCase.GetTenantUsersByTenant(cmd)
	if users == nil {
		users = make([]User, 0)
	}
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetTenantAdmins(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding URI
	var uriDto dto.TenantUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// Binding Query
	queryDto := GetUserListQueryDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// Esecuzione comando
	cmd := GetTenantAdminsByTenantCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
		TenantId:  tenantId,
	}

	users, total, err := controller.getTenantAdminsByTenantUseCase.GetTenantAdminsByTenant(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		} else if errors.Is(err, ErrUserNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *Controller) GetSuperAdmins(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// Binding Query
	queryDto := GetUserListQueryDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	// Esecuzione comando
	cmd := GetSuperAdminListCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
	}

	users, total, err := controller.getSuperAdminListUseCase.GetSuperAdminList(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// Invio risposta
	responseDto := NewUserListResponseDTO(users, total)
	ctx.JSON(http.StatusOK, responseDto)
}
