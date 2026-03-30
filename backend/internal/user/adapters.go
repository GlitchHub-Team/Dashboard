package user

import (
	"backend/internal/infra/database"
	"backend/internal/infra/database/pagination"

	"backend/internal/shared/identity"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/user/mocks/ports.go -package=mocks . SaveUserPort,DeleteUserPort,GetUserPort

type UserPostgreAdapter struct {
	log              *zap.Logger
	tenantMemberRepo TenantMemberRepository
	superAdminRepo   SuperAdminRepository
}

func NewUserPostgreAdapter(
	log *zap.Logger,
	tenantMemberRepo TenantMemberRepository,
	superAdminRepo SuperAdminRepository,
) *UserPostgreAdapter {
	adapter := &UserPostgreAdapter{
		log:              log,
		tenantMemberRepo: tenantMemberRepo,
		superAdminRepo:   superAdminRepo,
	}
	return adapter
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

func (adapter *UserPostgreAdapter) SaveUser(user User) (
	savedUser User, err error,
) {
	switch user.Role {
	case identity.ROLE_TENANT_USER, identity.ROLE_TENANT_ADMIN:
		entity := UserToTenantMemberEntity(user)

		err = adapter.tenantMemberRepo.SaveTenantMember(entity)
		if err != nil {
			return
		}

		savedUser, err = TenantMemberEntityToUser(entity)
		return

	case identity.ROLE_SUPER_ADMIN:
		entity := UserToSuperAdminEntity(user)

		err = adapter.superAdminRepo.SaveSuperAdmin(entity)
		if err != nil {
			return
		}

		savedUser, err = SuperAdminEntityToUser(entity)
		return

	default:
		err = identity.ErrUnknownRole
		return
	}
}

// Delete =============================================================================================
type DeleteUserPort interface {
	DeleteTenantUser(tenantId uuid.UUID, userId uint) (User, error)
	DeleteTenantAdmin(tenantId uuid.UUID, userId uint) (User, error)
	DeleteSuperAdmin(userId uint) (User, error)
}

func (adapter *UserPostgreAdapter) deleteTenantMember(tenantId uuid.UUID, userId uint) (
	user User, err error,
) {
	oldMember := TenantMemberEntity{
		ID:       userId,
		TenantId: tenantId.String(),
	}
	err = adapter.tenantMemberRepo.DeleteTenantMember(&oldMember)
	if err != nil {
		return User{}, err
	}
	user, err = TenantMemberEntityToUser(&oldMember)
	return
}

func (adapter *UserPostgreAdapter) DeleteTenantUser(tenantId uuid.UUID, userId uint) (
	user User, err error,
) {
	user, err = adapter.deleteTenantMember(tenantId, userId)
	return
}

func (adapter *UserPostgreAdapter) DeleteTenantAdmin(tenantId uuid.UUID, userId uint) (
	user User, err error,
) {
	user, err = adapter.deleteTenantMember(tenantId, userId)
	return
}

func (adapter *UserPostgreAdapter) DeleteSuperAdmin(userId uint) (user User, err error) {
	oldMember := SuperAdminEntity{
		ID: userId,
	}
	err = adapter.superAdminRepo.DeleteSuperAdmin(&oldMember)
	if err != nil {
		return User{}, err
	}
	user, _ = SuperAdminEntityToUser(&oldMember)
	return
}

// Get ================================================================================================
// TODO: fix what happens if i pass zero tenant id in GetTenantUser/-Admin (normal and by email)
type GetUserPort interface {
	GetUser(tenantId *uuid.UUID, userId uint) (User, error)

	GetUserByEmail(tenantId *uuid.UUID, email string) (User, error)

	GetTenantUsersByTenant(tenantId uuid.UUID, page, limit int) (
		tenantUsers []User, total uint, err error,
	)
	GetTenantAdminsByTenant(tenantId uuid.UUID, page, limit int) (
		tenantAdmins []User, total uint, err error,
	)
	GetSuperAdminList(page, limit int) (
		superAdmins []User, total uint, err error,
	)

	CountTenantAdminsByTenant(tenantId uuid.UUID) (total uint, err error)
	CountSuperAdmins() (total uint, err error)
}

func (adapter *UserPostgreAdapter) GetUser(tenantId *uuid.UUID, userId uint) (
	user User, err error,
) {
	getUserBy := UserRepositoryGetUserBy{UserId: &userId}
	// Tenant User / Tenant Admin
	if tenantId != nil {
		tenantMember, err := adapter.tenantMemberRepo.GetTenantMember(
			tenantId.String(), getUserBy,
		)
		if err != nil {
			return User{}, err
		}

		user, err = TenantMemberEntityToUser(tenantMember)

		return user, err
	}

	// Super Admin
	superAdmin, err := adapter.superAdminRepo.GetSuperAdmin(getUserBy)
	if err != nil {
		return User{}, err
	}

	user, err = SuperAdminEntityToUser(superAdmin)

	return user, err
}

func (adapter *UserPostgreAdapter) GetUserByEmail(tenantId *uuid.UUID, email string) (user User, err error) {
	// Tenant User / Tenant Admin
	if tenantId != nil {
		tenantMember, err := adapter.tenantMemberRepo.GetTenantMember(
			tenantId.String(),
			UserRepositoryGetUserBy{Email: &email},
		)
		if err != nil {
			return User{}, err
		}

		user, err = TenantMemberEntityToUser(tenantMember)
		return user, err
	}

	// Super Admin
	superAdmin, err := adapter.superAdminRepo.GetSuperAdmin(UserRepositoryGetUserBy{Email: &email})
	if err != nil {
		return User{}, err
	}

	user, err = SuperAdminEntityToUser(superAdmin)

	return
}

func (adapter *UserPostgreAdapter) GetTenantUsersByTenant(tenantId uuid.UUID, page, limit int) (
	tenantUsers []User, total uint, err error,
) {
	tenantUsers = make([]User, 0)
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return
	}

	entities, tot, err := adapter.tenantMemberRepo.GetTenantUsers(tenantId.String(), offset, limit)
	total = uint(tot)
	if err != nil {
		return
	}

	tenantUsers, err = database.MapEntityListToDomain(entities, TenantMemberEntityToUser)
	return
}

func (adapter *UserPostgreAdapter) GetTenantAdminsByTenant(tenantId uuid.UUID, page, limit int) (
	tenantAdmins []User, total uint, err error,
) {
	tenantAdmins = make([]User, 0)
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return
	}

	entities, tot, err := adapter.tenantMemberRepo.GetTenantAdmins(tenantId.String(), offset, limit)
	total = uint(tot)
	if err != nil {
		return
	}

	tenantAdmins, err = database.MapEntityListToDomain(entities, TenantMemberEntityToUser)
	return
}

func (adapter *UserPostgreAdapter) GetSuperAdminList(page, limit int) (
	superAdmins []User, total uint, err error,
) {
	superAdmins = make([]User, 0)
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return
	}

	entities, tot, err := adapter.superAdminRepo.GetSuperAdmins(offset, limit)
	total = uint(tot)
	if err != nil {
		return
	}

	superAdmins, err = database.MapEntityListToDomain(entities, SuperAdminEntityToUser)
	return
}

func (adapter *UserPostgreAdapter) CountTenantAdminsByTenant(tenantId uuid.UUID) (total uint, err error) {
	tot, err := adapter.tenantMemberRepo.CountTenantAdminsByTenant(tenantId.String())
	total = uint(tot)
	if err != nil {
		return 0, err
	}
	return
}

func (adapter *UserPostgreAdapter) CountSuperAdmins() (total uint, err error) {
	tot, err := adapter.superAdminRepo.CountSuperAdmins()
	total = uint(tot)
	if err != nil {
		return 0, err
	}
	return
}
