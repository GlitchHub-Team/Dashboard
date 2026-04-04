package user

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

/*
Mappa struct *TenantMemberEntity a User.
*/
func TenantMemberEntityToUser(entity *TenantMemberEntity) (user User, err error) {
	if entity == nil || entity.ID == uint(0) {
		return
	}

	tenantId, err := uuid.Parse(entity.TenantId)
	if err != nil {
		return
	}

	return User{
		Id:           entity.ID,
		Name:         entity.Name,
		Email:        entity.Email,
		PasswordHash: entity.Password,
		Role:         identity.UserRole(entity.Role),
		TenantId:     &tenantId,
		Confirmed:    entity.Confirmed,
	}, nil
}

/*
Mappa struct User a *TenantMemberEntity
*/
func UserToTenantMemberEntity(user User) *TenantMemberEntity {
	return &TenantMemberEntity{
		ID:        user.Id,
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.PasswordHash,
		Confirmed: user.Confirmed,
		Role:      string(user.Role),
		TenantId:  user.TenantId.String(),
	}
}

/*
Mappa struct *SuperAdminEntity a User.

NOTA: ritorna sempre errore nil, ma serve per essere compatibile con backend/internal/infra/database.MapEntityListToDomain()
*/
func SuperAdminEntityToUser(entity *SuperAdminEntity) (user User, err error) {
	if entity == nil || entity.ID == uint(0)  {
		return
	}
	user = User{
		Id:           entity.ID,
		Name:         entity.Name,
		Email:        entity.Email,
		PasswordHash: entity.Password,
		Role:         identity.ROLE_SUPER_ADMIN,
		TenantId:     nil,
		Confirmed:    entity.Confirmed,
	}
	return
}

/*
Mappa struct User a *SuperAdminEntity.
*/
func UserToSuperAdminEntity(user User) *SuperAdminEntity {
	return &SuperAdminEntity{
		ID:        user.Id,
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.PasswordHash,
		Confirmed: user.Confirmed,
	}
}
