package user

import (
	"backend/internal/infra/database/pagination"

	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/user/mocks/ports.go -package=mocks . SaveUserPort,DeleteUserPort,GetUserPort

type UserPostgreAdapter struct {
	repo *userPostgreRepository
}

func NewUserPostgreAdapter(repository *userPostgreRepository) (
	SaveUserPort,
	DeleteUserPort,
	GetUserPort,
) {
	adapter := &UserPostgreAdapter{
		repo: repository,
	}
	return adapter, adapter, adapter
}

// Compile-time checks
var (
	_ SaveUserPort   = (*UserPostgreAdapter)(nil)
	_ DeleteUserPort = (*UserPostgreAdapter)(nil)
	_ GetUserPort    = (*UserPostgreAdapter)(nil)
)

// Create =============================================================================================
type SaveUserPort interface {
	SaveUser(user User) (User, error)
}

func (adapter *UserPostgreAdapter) SaveUser(user User) (User, error) {
	switch user.Role {
	case identity.ROLE_TENANT_USER, identity.ROLE_TENANT_ADMIN:

		tenantMember := &TenantMemberEntity{}
		tenantMember.fromUser(user)
		err := adapter.repo.SaveTenantMember(tenantMember)
		if err != nil {
			return User{}, err
		}

		user, err := tenantMember.toUser()
		return user, err

	case identity.ROLE_SUPER_ADMIN:
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
		userRepositoryGetUserBy{UserId: &userId},
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
		userRepositoryGetUserBy{UserId: &userId},
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
		userRepositoryGetUserBy{UserId: &userId},
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
	GetUser(tenantId *uuid.UUID, userId uint) (User, error)
	GetTenantUser(tenantId uuid.UUID, userId uint) (User, error)
	GetTenantAdmin(tenantId uuid.UUID, userId uint) (User, error)
	GetSuperAdmin(userId uint) (User, error)

	GetUserByEmail(tenantId *uuid.UUID, email string) (User, error)
	GetTenantUserByEmail(tenantId uuid.UUID, email string) (User, error)
	GetTenantAdminByEmail(tenantId uuid.UUID, email string) (User, error)
	GetSuperAdminByEmail(email string) (User, error)
	

	GetTenantUsersByTenant(tenantId uuid.UUID, page, limit int) (
		tenantUsers []User, total uint, err error,
	)
	GetTenantAdminsByTenant(tenantId uuid.UUID, page, limit int) (
		tenantAdmins []User, total uint, err error,
	)
	GetSuperAdminList(page, limit int) (
		superAdmins []User, total uint, err error,
	)
}


func (adapter *UserPostgreAdapter) GetUser(tenantId *uuid.UUID, userId uint) (User, error) {
	var user User
	getUserBy := userRepositoryGetUserBy{UserId: &userId, }
	// Tenant User / Tenant Admin
	if tenantId != nil {
		tenantMember, err := adapter.repo.GetTenantMember(
			tenantId.String(), getUserBy,
		)
		if err != nil {
			return User{}, err
		}

		if *tenantMember != (TenantMemberEntity{}) {
			user, err = tenantMember.toUser()
		} else {
			user = User{}
		}

		return user, err
	}  

	// Super Admin
	superAdmin, err := adapter.repo.GetSuperAdmin(getUserBy)
	if err != nil {
		return User{}, err
	}

	if *superAdmin != (SuperAdminEntity{}) {
		user = superAdmin.toUser()
	} else {
		user = User{}
	}

	return user, err
}

func (adapter *UserPostgreAdapter) GetTenantUser(tenantId uuid.UUID, userId uint) (User, error) {
	tenantUser, err := adapter.repo.GetTenantUser(
		tenantId.String(),
		userRepositoryGetUserBy{UserId: &userId},
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
		userRepositoryGetUserBy{UserId: &userId},
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
		userRepositoryGetUserBy{UserId: &userId},
	)
	if err != nil {
		return User{}, err
	}
	user := superAdmin.toUser()
	return user, err
}

func (adapter *UserPostgreAdapter) GetUserByEmail(tenantId *uuid.UUID, email string) (User, error,) {
	var user User
	// Tenant User / Tenant Admin
	if tenantId != nil {
		tenantMember, err := adapter.repo.GetTenantMember(
			tenantId.String(), 
			userRepositoryGetUserBy{Email: &email,},
		)
		if err != nil {
			return User{}, err
		}

		if *tenantMember != (TenantMemberEntity{}) {
			user, err = tenantMember.toUser()
		} else {
			user = User{}
		}

		return user, err
	}  

	// Super Admin
	superAdmin, err := adapter.repo.GetSuperAdmin(userRepositoryGetUserBy{Email: &email,})
	if err != nil {
		return User{}, err
	}

	if *superAdmin != (SuperAdminEntity{}) {
		user = superAdmin.toUser()
	} else {
		user = User{}
	}

	return user, err
}

func (adapter *UserPostgreAdapter) GetTenantUserByEmail(tenantId uuid.UUID, email string) (User, error) {
	tenantUser, err := adapter.repo.GetTenantUser(
		tenantId.String(),
		userRepositoryGetUserBy{Email: &email},
	)
	if err != nil {
		return User{}, err
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

func (adapter *UserPostgreAdapter) GetTenantUsersByTenant(
	tenantId uuid.UUID, page, limit int,
) (tenantUsers []User, total uint, err error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, total, err := adapter.repo.GetTenantUsers(tenantId.String(), offset, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, entity := range entities {
		tenantUser, err := entity.toUser()
		if err != nil {
			return nil, 0, err
		}
		tenantUsers = append(tenantUsers, tenantUser)
	}

	return tenantUsers, total, nil
}

func (adapter *UserPostgreAdapter) GetTenantAdminsByTenant(
	tenantId uuid.UUID, page, limit int,
) (tenantAdmins []User, total uint, err error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, total, err := adapter.repo.GetTenantAdmins(tenantId.String(), offset, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, entity := range entities {
		tenantUser, err := entity.toUser()
		if err != nil {
			return nil, 0, err
		}
		tenantAdmins = append(tenantAdmins, tenantUser)
	}

	return tenantAdmins, total, nil
}

func (adapter *UserPostgreAdapter) GetSuperAdminList(page, limit int) (
	superAdmins []User, total uint, err error,
) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, total, err := adapter.repo.GetSuperAdmins(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, entity := range entities {
		tenantUser := entity.toUser()
		superAdmins = append(superAdmins, tenantUser)
	}
	return superAdmins, total, nil
}
