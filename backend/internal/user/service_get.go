package user


type GetTenantUserUseCase interface {
	GetTenantUser(cmd GetTenantUserCommand) (User, error)
}

type GetTenantAdminUseCase interface {
	GetTenantAdmin(cmd GetTenantAdminCommand) (User, error)
}

type GetSuperAdminUseCase interface {
	GetSuperAdmin(cmd GetSuperAdminCommand) (User, error)
}

type GetUsersUseCase interface {
	GetUsers(cmd GetUsersCommand) (users []User, total int, err error)
}

type GetUsersByTenantIdUseCase interface {
	GetUsersByTenantId(cmd GetUsersByTenantIdCommand) (users []User, total int, err error)
}


type GetUserService struct {
	getUserPort GetUserPort
}

func NewGetUserService(getUserPort GetUserPort) (
	GetTenantUserUseCase,
	GetTenantAdminUseCase,
	GetSuperAdminUseCase,
	GetUsersUseCase,
	GetUsersByTenantIdUseCase,
) {
	service := &GetUserService{
		getUserPort: getUserPort,
	}
	return service, service, service, service, service
}


func (service *GetUserService) GetTenantUser(cmd GetTenantUserCommand) (User, error) {
	user, err := service.getUserPort.GetTenantUser(cmd.TenantId, cmd.UserId)
	return user, err
}

func (service *GetUserService) GetTenantAdmin(cmd GetTenantAdminCommand) (User, error) {
	user, err := service.getUserPort.GetTenantAdmin(cmd.TenantId, cmd.UserId)
	return user, err
}

func (service *GetUserService) GetSuperAdmin(cmd GetSuperAdminCommand) (User, error) {
	user, err := service.getUserPort.GetSuperAdmin(cmd.UserId)
	return user, err
}

func (service *GetUserService) GetUsers(cmd GetUsersCommand) ([]User, int, error) {
	user, total, err := service.getUserPort.GetUsers(cmd.Page, cmd.Limit)
	return user, total, err
}

func (service *GetUserService) GetUsersByTenantId(cmd GetUsersByTenantIdCommand) ([]User, int, error) {
	user, total, err := service.getUserPort.GetUsersByTenantId(cmd.TenantId)
	return user, total, err
}

// Compile-time checks
var (
	_ GetTenantUserUseCase      = (*GetUserService)(nil)
	_ GetTenantAdminUseCase     = (*GetUserService)(nil)
	_ GetSuperAdminUseCase      = (*GetUserService)(nil)
	_ GetUsersUseCase           = (*GetUserService)(nil)
	_ GetUsersByTenantIdUseCase = (*GetUserService)(nil)
)
