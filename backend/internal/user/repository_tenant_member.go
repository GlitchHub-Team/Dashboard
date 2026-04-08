package user

import (
	"errors"
	"strings"
	"time"

	clouddb "backend/internal/infra/database/cloud_db/connection"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -destination=../../tests/user/mocks/repository_tenant_member.go -package=mocks . TenantMemberRepository

// Entity ===========================================================================================

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

// Repository =========================================================================================

type TenantMemberRepository interface {
	SaveTenantMember(tenantMember *TenantMemberEntity) error

	DeleteTenantMember(tenantMember *TenantMemberEntity) error

	GetTenantMember(tenantId string, by UserRepositoryGetUserBy) (
		tenantMember *TenantMemberEntity, err error,
	)
	// GetTenantUser(tenantId string, by UserRepositoryGetUserBy) (
	// 	tenantMember *TenantMemberEntity, err error,
	// )
	// GetTenantAdmin(tenantId string, by UserRepositoryGetUserBy) (
	// 	tenantMember *TenantMemberEntity, err error,
	// )

	GetTenantUsers(tenantId string, offset, limit int) (
		tenantAdmins []TenantMemberEntity, total int64, err error,
	)
	GetTenantAdmins(tenantId string, offset, limit int) (
		tenantAdmins []TenantMemberEntity, total int64, err error,
	)

	CountTenantAdminsByTenant(tenantId string) (total int64, err error)
}

type tenantMemberPgRepository struct {
	log *zap.Logger
	db  clouddb.CloudDBConnection
}

var _ TenantMemberRepository = (*tenantMemberPgRepository)(nil) // Compile-time check

func newTenantMemberPgRepository(
	log *zap.Logger,
	db clouddb.CloudDBConnection,
) *tenantMemberPgRepository {
	return &tenantMemberPgRepository{
		log: log,
		db:  db,
	}
}

// Save -------------------------------------------------------------------------------------------

func (repo *tenantMemberPgRepository) SaveTenantMember(tenantMember *TenantMemberEntity) error {
	db := (*gorm.DB)(repo.db)
	err := db.
		Scopes(clouddb.WithTenantSchema(tenantMember.TenantId, &TenantMemberEntity{})).
		Save(tenantMember).
		Error
	return err
}

// Delete -----------------------------------------------------------------------------------------

func (repo *tenantMemberPgRepository) DeleteTenantMember(tenantMember *TenantMemberEntity) error {
	if tenantMember.ID == 0 {
		return errors.New("cannot delete tenant member with ID 0")
	}

	if tenantMember.TenantId == "" {
		return errors.New("cannot delete tenant member with no tenant")
	}

	db := (*gorm.DB)(repo.db)
	err := db.
		Scopes(clouddb.WithTenantSchema(tenantMember.TenantId, &TenantMemberEntity{})).
		Clauses(clause.Returning{}).
		Delete(&tenantMember).
		Error
	return err
}

// Get singolo ------------------------------------------------------------------------------------------------------

type UserRepositoryGetUserBy struct {
	Email  *string
	UserId *uint
}

func (by *UserRepositoryGetUserBy) getWhere() (string, []interface{}, error) {
	if by == (&UserRepositoryGetUserBy{}) {
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

func (repo *tenantMemberPgRepository) GetTenantMember(tenantId string, by UserRepositoryGetUserBy) (
	tenantMember *TenantMemberEntity, err error,
) {
	where, params, err := by.getWhere()
	if err != nil {
		return
	}

	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where(where, params...).
		Find(tenantMember).
		Error

	if tenantMember != nil {
		tenantMember.TenantId = tenantId
	}
	return
}

func (repo *tenantMemberPgRepository) GetTenantUser(tenantId string, by UserRepositoryGetUserBy) (
	tenantMember *TenantMemberEntity, err error,
) {
	where, params, err := by.getWhere()
	if err != nil {
		return &TenantMemberEntity{}, err
	}

	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_user").
		Where(where, params...).
		Find(tenantMember).
		Error

	if tenantMember != nil {
		tenantMember.TenantId = tenantId
	}
	return
}

func (repo *tenantMemberPgRepository) GetTenantAdmin(tenantId string, by UserRepositoryGetUserBy) (
	tenantMember *TenantMemberEntity, err error,
) {
	where, params, err := by.getWhere()
	if err != nil {
		return
	}

	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_admin").
		Where(where, params...).
		Find(tenantMember).
		Error

	if tenantMember != nil {
		tenantMember.TenantId = tenantId
	}

	return
}

// Get multiplo ---------------------------------------------------------------------------------------------------------

func (repo *tenantMemberPgRepository) GetTenantUsers(tenantId string, offset, limit int) (
	tenantUsers []TenantMemberEntity, count int64, err error,
) {
	db := (*gorm.DB)(repo.db)
	baseQuery := db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantMemberEntity{})).
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

	return tenantUsers, count, nil
}

func (repo *tenantMemberPgRepository) GetTenantAdmins(tenantId string, offset, limit int) (
	tenantAdmins []TenantMemberEntity, total int64, err error,
) {
	db := (*gorm.DB)(repo.db)
	baseQuery := db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_admin")

	// Ottieni count
	if err := baseQuery.Count(&total).Error; err != nil {
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

	return
}

// Conteggio ------------------------------------------------------------------------------------------

func (repo *tenantMemberPgRepository) CountTenantAdminsByTenant(tenantId string) (total int64, err error) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantMemberEntity{})).
		Where("role = ?", "tenant_admin").
		Count(&total).
		Error

	return
}
