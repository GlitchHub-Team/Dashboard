package user

import (
	"errors"
	"strings"
	"time"

	cloud_db "backend/internal/infra/database/cloud_db/connection"

	"backend/internal/shared/identity"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Entities ===========================================================================================

type TenantMemberEntity struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique;size:256;not null"`
	Name      string `gorm:"size:128;not null"`
	Password  *string
	Confirmed bool   `gorm:"not null"`
	Role      string `gorm:"not null;size:32;check:role = 'tenant_user' or role ='tenant_admin'"`

	// NOTA: Questo parametro è ignorato, ma è FONDAMENTALE perché va specificato quando si chiama il repository!
	TenantId  string `gorm:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TenantMemberEntity) TableName() string { return "tenant_members" }

func (entity *TenantMemberEntity) fromUser(user User) {
	entity.Email = user.Email
	entity.Name = user.Name
	entity.Password = user.PasswordHash
	entity.Role = string(user.Role)
	entity.TenantId = user.TenantId.String()
	entity.Confirmed = user.Confirmed
}

func (entity *TenantMemberEntity) toUser() (User, error) {
	if entity.ID == 0 {
		return User{}, ErrInvalidUser
	}

	tenantId, err := uuid.Parse(entity.TenantId)
	var tenantIdPointer *uuid.UUID
	if err != nil {
		tenantIdPointer = nil
	} else {
		tenantIdPointer = &tenantId
	}

	return User{
		Id:           entity.ID,
		Name:         entity.Name,
		Email:        entity.Email,
		PasswordHash: entity.Password,
		Role:         (identity.UserRole)(entity.Role),
		TenantId:     tenantIdPointer,
		Confirmed:    entity.Confirmed,
	}, err
}

type SuperAdminEntity struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique;size:256;not null"`
	Name      string `gorm:"size:128;not null"`
	Password  *string
	CreatedAt time.Time
	UpdatedAt time.Time
	Confirmed bool `gorm:"not null"`
}

func (SuperAdminEntity) TableName() string { return "super_admins" }

func (entity *SuperAdminEntity) fromUser(user User) *SuperAdminEntity {
	entity.ID = user.Id
	entity.Email = user.Email
	entity.Name = user.Name
	entity.Password = user.PasswordHash
	return entity
}

func (entity *SuperAdminEntity) toUser() User {
	return User{
		Id:           entity.ID,
		Name:         entity.Name,
		Email:        entity.Email,
		PasswordHash: entity.Password,
		Confirmed:    entity.Confirmed,
	}
}

// Repository =========================================================================================
type userPostgreRepository struct {
	log *zap.Logger
	db  *gorm.DB
}

func newUserPostgreRepository(
	log *zap.Logger,
	db *gorm.DB,
) *userPostgreRepository {
	return &userPostgreRepository{
		log: log,
		db:  db,
	}
}

func (repo *userPostgreRepository) SaveTenantMember(tenantMember *TenantMemberEntity) error {
	err := repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantMember.TenantId, &TenantMemberEntity{})).
		Save(tenantMember).
		Error
	return err
}

func (repo *userPostgreRepository) DeleteTenantMember(tenantMember *TenantMemberEntity) error {
	if tenantMember.ID == 0 {
		return errors.New("cannot delete tenant member with ID 0")
	}

	if tenantMember.TenantId == "" {
		return errors.New("cannot delete tenant member with no tenant")
	}

	err := repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantMember.TenantId, &TenantMemberEntity{})).
		Clauses(clause.Returning{}).
		Delete(&tenantMember).
		Error
	return err
}

type userRepositoryGetUserBy struct {
	Email  *string
	UserId *uint
}

func (by *userRepositoryGetUserBy) getWhere() (string, []interface{}, error) {
	if by == (&userRepositoryGetUserBy{}) {
		return "", nil, errors.New("cannot get user without specifying parameters")
	}

	var conditions []string
	var params []any
	if by.Email != nil {
		conditions = append(conditions, "email = ?")
		params = append(params, *by.Email)
	}
	if by.UserId != nil {
		conditions = append(conditions, "id = ?")
		params = append(params, *by.UserId)
	}

	return strings.Join(conditions, " AND "), params, nil
}

func (repo *userPostgreRepository) GetTenantMember(tenantId string, by userRepositoryGetUserBy) (*TenantMemberEntity, error) {
	where, params, err := by.getWhere()
	if err != nil {
		return &TenantMemberEntity{}, err
	}

	var tenantMember *TenantMemberEntity

	err = repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where(where, params...).
		Find(&tenantMember).
		Error
	tenantMember.TenantId = tenantId
	return tenantMember, err
}

func (repo *userPostgreRepository) GetTenantUser(tenantId string, by userRepositoryGetUserBy) (*TenantMemberEntity, error) {
	where, params, err := by.getWhere()
	if err != nil {
		return &TenantMemberEntity{}, err
	}

	var tenantMember *TenantMemberEntity

	err = repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_user").
		Where(where, params...).
		Find(&tenantMember).
		Error
	tenantMember.TenantId = tenantId
	return tenantMember, err
}

func (repo *userPostgreRepository) GetTenantUsers(tenantId string, offset, limit int) ([]TenantMemberEntity, uint, error) {
	var tenantUsers []TenantMemberEntity
	var count int64
	var err error

	baseQuery := repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_user")

	// Get count
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get tenant users
	err = baseQuery.Order("name ASC").Limit(limit).Offset(offset).Find(&tenantUsers).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range tenantUsers {
		tenantUsers[i].TenantId = tenantId
	}

	return tenantUsers, uint(count), nil
}

func (repo *userPostgreRepository) GetTenantAdmin(tenantId string, by userRepositoryGetUserBy) (*TenantMemberEntity, error) {
	var tenantMember *TenantMemberEntity
	where, params, err := by.getWhere()
	if err != nil {
		return &TenantMemberEntity{}, err
	}

	err = repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_admin").
		Where(where, params...).
		Find(&tenantMember).
		Error
	tenantMember.TenantId = tenantId
	return tenantMember, err
}

func (repo *userPostgreRepository) GetTenantAdmins(tenantId string, offset, limit int) ([]TenantMemberEntity, uint, error) {
	var tenantAdmins []TenantMemberEntity
	var count int64
	var err error

	baseQuery := repo.db.
		Scopes(cloud_db.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_admin")

	// Ottieni count
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Ottieni tenant admins
	err = baseQuery.Order("name ASC").Limit(limit).Offset(offset).Find(&tenantAdmins).Error
	if err != nil {
		return nil, 0, err
	}

	// Assegna tenantId
	for i := range tenantAdmins {
		tenantAdmins[i].TenantId = tenantId
	}

	return tenantAdmins, uint(count), nil
}

func (repo *userPostgreRepository) SaveSuperAdmin(tenantMember *SuperAdminEntity) error {
	err := repo.db.Save(tenantMember).Error
	return err
}

func (repo *userPostgreRepository) DeleteSuperAdmin(tenantMember *SuperAdminEntity) error {
	if tenantMember.ID == 0 {
		return errors.New("cannot delete super admin with ID 0")
	}
	err := repo.db.
		Clauses(clause.Returning{}).
		Delete(tenantMember).
		Error
	return err
}

func (repo *userPostgreRepository) GetSuperAdmin(by userRepositoryGetUserBy) (*SuperAdminEntity, error) {
	var tenantMember *SuperAdminEntity
	where, params, err := by.getWhere()
	if err != nil {
		return &SuperAdminEntity{}, err
	}

	err = repo.db.
		Where(where, params...).
		Find(&tenantMember).
		Error
	return tenantMember, err
}

func (repo *userPostgreRepository) GetSuperAdmins(offset, limit int) ([]SuperAdminEntity, uint, error) {
	var superAdmins []SuperAdminEntity
	var count int64
	var err error

	baseQuery := repo.db.
		Model(&SuperAdminEntity{}).
		Where("role = ?", "tenant_user")

	// Ottieni count
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Ottieni super admin
	err = baseQuery.Order("name ASC").Limit(limit).Offset(offset).Find(&superAdmins).Error
	if err != nil {
		return nil, 0, err
	}

	return superAdmins, uint(count), nil
}
