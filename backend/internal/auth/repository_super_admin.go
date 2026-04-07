package auth

import (
	"time"

	"backend/internal/user"

	clouddb "backend/internal/infra/database/cloud_db/connection"

	"gorm.io/gorm"
)

/*
Entity per token di conferma account all'interno per un super admin
*/
type SuperAdminConfirmTokenEntity struct {
	Token string `gorm:"primaryKey;index:,type:hash"`

	UserId     uint                  `gorm:"not null"`
	SuperAdmin user.SuperAdminEntity `gorm:"foreignKey:UserId;not null"`

	CreatedAt time.Time
	ExpiresAt time.Time
}

func ConfirmAccountTokenToSuperAdminEntity(tokenObj ConfirmAccountToken) *SuperAdminConfirmTokenEntity {
	return &SuperAdminConfirmTokenEntity{
		Token:     tokenObj.Token,
		UserId:    tokenObj.UserId,
		ExpiresAt: tokenObj.ExpiryDate,
	}
}

func SuperAdminConfirmTokenEntityToConfirmAccountToken(entity *SuperAdminConfirmTokenEntity) ConfirmAccountToken {
	return ConfirmAccountToken{
		Token: entity.Token,
		UserId:      entity.UserId,
		ExpiryDate:  entity.ExpiresAt,
	}
}

func (SuperAdminConfirmTokenEntity) TableName() string { return "super_admin_confirm_tokens" }

// repository -----------------------------------------------------------------------------------------

type superAdminConfirmTokenPgRepository struct {
	db clouddb.CloudDBConnection
}

var _ SuperAdminConfirmTokenRepository = (*superAdminConfirmTokenPgRepository)(nil)

func newSuperAdminConfirmTokenPgRepository(db clouddb.CloudDBConnection) *superAdminConfirmTokenPgRepository {
	return &superAdminConfirmTokenPgRepository{
		db: db,
	}
}

func (repo *superAdminConfirmTokenPgRepository) SaveToken(entity *SuperAdminConfirmTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.Save(entity).Error
	return
}

func (repo *superAdminConfirmTokenPgRepository) DeleteToken(entity *SuperAdminConfirmTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.Delete(entity).Error
	return
}

func (repo *superAdminConfirmTokenPgRepository) GetToken(tokenString string) (
	entity *SuperAdminConfirmTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Where("token = ?", tokenString).
		Find(&entity).
		Error
	return
}

func (repo *superAdminConfirmTokenPgRepository) GetTokenWithUser(tokenString string) (
	entity *SuperAdminConfirmTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Preload("SuperAdmin").
		Where("token = ?", tokenString).
		Find(&entity).
		Error

	if err != nil {
		return
	}

	err = db.
		Find(&entity.SuperAdmin, entity.UserId).
		Error

	return
}

// Forgot Password ====================================================================================

/*
Entity per i token di cambio password dimenticata per un super admin
*/
type SuperAdminPasswordTokenEntity struct {
	Token string `gorm:"primaryKey;index:,type:hash"`

	UserId     uint                  `gorm:"not null"`
	SuperAdmin user.SuperAdminEntity `gorm:"foreignKey:UserId;not null"`

	CreatedAt time.Time
	ExpiresAt time.Time
}

func ForgotPasswordTokenToSuperAdminTokenEntity(tokenObj ForgotPasswordToken) *SuperAdminPasswordTokenEntity {
	return &SuperAdminPasswordTokenEntity{
		Token:     tokenObj.Token,
		UserId:    tokenObj.UserId,
		ExpiresAt: tokenObj.ExpiryDate,
	}
}

func SuperAdminPasswordTokenEntityToForgotPasswordToken(entity *SuperAdminPasswordTokenEntity) ForgotPasswordToken {
	return ForgotPasswordToken{
		Token: entity.Token,
		UserId:      entity.UserId,
		ExpiryDate:  entity.ExpiresAt,
	}
}

func (SuperAdminPasswordTokenEntity) TableName() string { return "super_admin_forgot_password_tokens" }

// repository -----------------------------------------------------------------------------------------

type superAdminPasswordTokenPgRepository struct {
	db clouddb.CloudDBConnection
}

var _ SuperAdminPasswordTokenRepository = (*superAdminPasswordTokenPgRepository)(nil)

func newSuperAdminPasswordTokenPgRepository(db clouddb.CloudDBConnection) *superAdminPasswordTokenPgRepository {
	return &superAdminPasswordTokenPgRepository{
		db: db,
	}
}

func (repo *superAdminPasswordTokenPgRepository) SaveToken(entity *SuperAdminPasswordTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.Save(entity).Error
	return
}

func (repo *superAdminPasswordTokenPgRepository) DeleteToken(entity *SuperAdminPasswordTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.Delete(entity).Error
	return
}

func (repo *superAdminPasswordTokenPgRepository) GetToken(tokenString string) (
	entity *SuperAdminPasswordTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Where("token = ?", tokenString).
		Find(&entity).
		Error
	return
}

func (repo *superAdminPasswordTokenPgRepository) GetTokenWithUser(tokenString string) (
	entity *SuperAdminPasswordTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Preload("SuperAdmin").
		Where("token = ?", tokenString).
		Find(&entity).
		Error

	if err != nil {
		return
	}

	err = db.
		Find(&entity.SuperAdmin, entity.UserId).
		Error

	return
}
