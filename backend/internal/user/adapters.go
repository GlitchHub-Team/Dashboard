package user

import (
	"fmt"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/user/mocks/ports.go -package=mocks . CreateUserPort,DeleteUserPort,GetUserPort

type UserPostgreAdapter struct {
	repo *userPostgreRepository
}

func NewUserPostgreAdapter(repository *userPostgreRepository) (
	CreateUserPort,
	DeleteUserPort,
	GetUserPort,
) {
	adapter := &UserPostgreAdapter{
		repo: repository,
	}
	return adapter, adapter, adapter
}

// Create =============================================================================================
type CreateUserPort interface {
	CreateUser(user User) (User, error)
}

func (adapter *UserPostgreAdapter) CreateUser(user User) (User, error) {
	switch user.Role {
	case ROLE_TENANT_USER, ROLE_TENANT_ADMIN:

		tenantMember := &TenantMemberEntity{}
		tenantMember.fromUser(user)
		err := adapter.repo.SaveTenantMember(tenantMember)
		if err != nil {
			return User{}, err
		}

		user, err := tenantMember.toUser()
		return user, err

	case ROLE_SUPER_ADMIN:
		superAdmin := (&SuperAdminEntity{}).fromUser(user)
		err := adapter.repo.SaveSuperAdmin(superAdmin)
		if err != nil {
			return User{}, err
		}

		user := superAdmin.toUser()
		return user, err

	default:
		return User{}, ErrUnknownRole
	}
}

// Delete =============================================================================================
type DeleteUserPort interface {
	DeleteTenantUser(tenantId uuid.UUID, userId uint) (User, error)
	DeleteTenantAdmin(tenantId uuid.UUID, userId uint) (User, error)
	DeleteSuperAdmin(userId uint) (User, error)
}

func (adapter *UserPostgreAdapter) DeleteTenantUser(tenantId uuid.UUID, userId uint) (User, error) {
	oldMember, err := adapter.repo.GetTenantUser(
		tenantId.String(),
		userRepositoryGetUserBy{userId: &userId},
	)
	if err != nil {
		return User{}, err
	}
	err = adapter.repo.DeleteTenantMember(oldMember)
	if err != nil {
		return User{}, err
	}
	user, err := oldMember.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) DeleteTenantAdmin(tenantId uuid.UUID, userId uint) (User, error) {
	oldMember, err := adapter.repo.GetTenantAdmin(
		tenantId.String(),
		userRepositoryGetUserBy{userId: &userId},
	)
	if err != nil {
		return User{}, err
	}
	err = adapter.repo.DeleteTenantMember(oldMember)
	if err != nil {
		return User{}, err
	}
	user, err := oldMember.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) DeleteSuperAdmin(userId uint) (User, error) {
	oldMember, err := adapter.repo.GetSuperAdmin(
		userRepositoryGetUserBy{userId: &userId},
	)
	if err != nil {
		return User{}, err
	}
	err = adapter.repo.DeleteSuperAdmin(oldMember)
	if err != nil {
		return User{}, err
	}
	user := oldMember.toUser()
	return user, err
}

// Get ================================================================================================
type GetUserPort interface {
	// GetUserById(userId uint) (*User, error)
	GetTenantUser(tenantId uuid.UUID, userId uint) (User, error)
	GetTenantAdmin(tenantId uuid.UUID, userId uint) (User, error)
	GetSuperAdmin(userId uint) (User, error)

	GetTenantUserByEmail(tenantId uuid.UUID, email string) (User, error)
	GetTenantAdminByEmail(tenantId uuid.UUID, email string) (User, error)
	GetSuperAdminByEmail(email string) (User, error)

	GetUsers(page, limit int) ([]User, int, error)
	GetUsersByRole(role UserRole, page, limit int) ([]User, int, error)
	GetUsersByTenantId(tenantId uuid.UUID) ([]User, int, error)
}

func (adapter *UserPostgreAdapter) GetTenantUser(tenantId uuid.UUID, userId uint) (User, error) {
	tenantUser, err := adapter.repo.GetTenantUser(
		tenantId.String(),
		userRepositoryGetUserBy{userId: &userId},
	)
	tenantUser.TenantId = tenantId.String()
	if err != nil {
		return User{}, err
	}
	user, err := tenantUser.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) GetTenantAdmin(tenantId uuid.UUID, userId uint) (User, error) {
	tenantAdmin, err := adapter.repo.GetTenantAdmin(
		tenantId.String(),
		userRepositoryGetUserBy{userId: &userId},
	)
	tenantAdmin.TenantId = tenantId.String()
	if err != nil {
		return User{}, err
	}
	user, err := tenantAdmin.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) GetSuperAdmin(userId uint) (User, error) {
	superAdmin, err := adapter.repo.GetSuperAdmin(
		userRepositoryGetUserBy{userId: &userId},
	)
	if err != nil {
		return User{}, err
	}
	user := superAdmin.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) GetTenantUserByEmail(tenantId uuid.UUID, email string) (User, error) {
	tenantUser, err := adapter.repo.GetTenantUser(
		tenantId.String(),
		userRepositoryGetUserBy{Email: &email},
	)
	if err != nil {
		return User{}, fmt.Errorf("error GetTenantUser: %v", err)
	}
	tenantUser.TenantId = tenantId.String()
	var user User

	if *tenantUser != (TenantMemberEntity{}) {
		user, err = tenantUser.toUser()
	} else {
		user, err = User{}, nil
	}
	return user, err
}

func (adapter *UserPostgreAdapter) GetTenantAdminByEmail(tenantId uuid.UUID, email string) (User, error) {
	tenantAdmin, err := adapter.repo.GetTenantAdmin(
		tenantId.String(),
		userRepositoryGetUserBy{Email: &email},
	)
	if err != nil {
		return User{}, err
	}

	tenantAdmin.TenantId = tenantId.String()
	user, err := tenantAdmin.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) GetSuperAdminByEmail(email string) (User, error) {
	superAdmin, err := adapter.repo.GetSuperAdmin(
		userRepositoryGetUserBy{Email: &email},
	)
	if err != nil {
		return User{}, err
	}
	user := superAdmin.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) GetUsers(page, limit int) ([]User, int, error) {
	return nil, 0, nil
}

func (adapter *UserPostgreAdapter) GetUsersByRole(role UserRole, page, limit int) ([]User, int, error) {
	return nil, 0, nil
}

func (adapter *UserPostgreAdapter) GetUsersByTenantId(tenantId uuid.UUID) ([]User, int, error) {
	return nil, 0, nil
}

// Compile-time checks
var (
	_ CreateUserPort = (*UserPostgreAdapter)(nil)
	_ DeleteUserPort = (*UserPostgreAdapter)(nil)
	_ GetUserPort    = (*UserPostgreAdapter)(nil)
)
