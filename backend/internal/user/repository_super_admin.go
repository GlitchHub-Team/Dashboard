package user

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	clouddb "backend/internal/infra/database/cloud_db/connection"
)

//go:generate mockgen -destination=../../tests/user/mocks/repository_super_admin.go -package=mocks . SuperAdminRepository

// Entity ===========================================================================================

type SuperAdminEntity struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique;size:256;not null"`
	Name      string `gorm:"size:128;not null"`
	Password  *string
	Confirmed bool `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SuperAdminEntity) TableName() string { return "super_admins" }

// Repository =========================================================================================

type SuperAdminRepository interface {
	SaveSuperAdmin(superAdmin *SuperAdminEntity) error
	DeleteSuperAdmin(superAdmin *SuperAdminEntity) error
	GetSuperAdmin(by UserRepositoryGetUserBy) (*SuperAdminEntity, error)
	GetSuperAdmins(offset, limit int) (
		superAdmins []SuperAdminEntity, total int64, err error,
	)
	CountSuperAdmins() (total int64, err error)
}

type superAdminPgRepository struct {
	// log *zap.Logger
	db clouddb.CloudDBConnection
}

var _ SuperAdminRepository = (*superAdminPgRepository)(nil) // Compile-time check

func newSuperAdminPgRepository(
	log *zap.Logger,
	db clouddb.CloudDBConnection,
) *superAdminPgRepository {
	return &superAdminPgRepository{
		db: db,
	}
}

// Save -------------------------------------------------------------------------------------------

func (repo *superAdminPgRepository) SaveSuperAdmin(superAdmin *SuperAdminEntity) error {
	db := (*gorm.DB)(repo.db)
	err := db.Save(superAdmin).Error
	return err
}

// Delete -----------------------------------------------------------------------------------------

func (repo *superAdminPgRepository) DeleteSuperAdmin(superAdmin *SuperAdminEntity) error {
	if superAdmin.ID == 0 {
		return errors.New("cannot delete super admin with ID 0")
	}
	db := (*gorm.DB)(repo.db)
	err := db.
		Clauses(clause.Returning{}).
		Delete(superAdmin).
		Error
	return err
}

// Get singolo ------------------------------------------------------------------------------------------------------

func (repo *superAdminPgRepository) GetSuperAdmin(by UserRepositoryGetUserBy) (*SuperAdminEntity, error) {
	var tenantMember *SuperAdminEntity
	where, params, err := by.getWhere()
	if err != nil {
		return &SuperAdminEntity{}, err
	}

	db := (*gorm.DB)(repo.db)
	err = db.
		Where(where, params...).
		Find(&tenantMember).
		Error
	return tenantMember, err
}

// Get multiplo ---------------------------------------------------------------------------------------------------------

func (repo *superAdminPgRepository) GetSuperAdmins(offset, limit int) (
	superAdmins []SuperAdminEntity, total int64, err error,
) {
	db := (*gorm.DB)(repo.db)
	baseQuery := db.
		Model(&SuperAdminEntity{})

	// Ottieni count
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Ottieni super admin
	err = baseQuery.Order("name ASC").Limit(limit).Offset(offset).Find(&superAdmins).Error
	if err != nil {
		return nil, 0, err
	}

	return
}

// Conteggio ---------------------------------------------------------------------------------------------------------

func (repo *superAdminPgRepository) CountSuperAdmins() (total int64, err error) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Model(&SuperAdminEntity{}).
		Count(&total).
		Error

	return
}
