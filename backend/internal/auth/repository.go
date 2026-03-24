package auth

import (
	"time"

	"backend/internal/tenant"
	"backend/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NOTA: Possibile miglioria è di inserire tabelle in schema per tenant, 
// invece che nel public per tutti i ruoli
type ConfirmTokenEntity struct {
	Token string `gorm:"primaryKey;index:,type:hash"`

	TenantId *string              `gorm:"size:36"`
	Tenant   tenant.TenantEntity `gorm:"foreignKey:TenantId;not null"`

	UserId uint                    `gorm:"not null"`
	User   user.TenantMemberEntity `gorm:"foreignKey:UserId;not null"`

	CreatedAt time.Time
	ExpiresAt time.Time
}

func newConfirmTokenEntityFromDomain(tokenObj ConfirmAccountToken) (ConfirmTokenEntity) {
	var tenantId *string
	if tokenObj.tenantId != nil {
		str := tokenObj.tenantId.String()
		tenantId = &str
	}
	return ConfirmTokenEntity{
		Token: tokenObj.hashedToken,
		TenantId: tenantId,
		UserId: tokenObj.userId,
		ExpiresAt: tokenObj.expiryDate,
	}
}

func (entity *ConfirmTokenEntity) ToConfirmToken() (ConfirmAccountToken, error) {
	tenantId, err := uuid.Parse(*entity.TenantId)
	if err != nil {
		return ConfirmAccountToken{}, err
	}
	return ConfirmAccountToken{
		hashedToken: entity.Token,
		tenantId: &tenantId,
		userId: entity.UserId,
		expiryDate: entity.ExpiresAt,
	}, nil
}

func (ConfirmTokenEntity) TableName() string { return "confirm_tokens" }

// repository -----------------------------------------------------------------------------------------

type confirmTokenPostgreRepository struct {
	db *gorm.DB
}

func newConfirmTokenPostgreRepository(db *gorm.DB) *confirmTokenPostgreRepository {
	return &confirmTokenPostgreRepository{
		db: db,
	}
}

func (repo *confirmTokenPostgreRepository) SaveToken(entity *ConfirmTokenEntity) (err error,) {
	err = repo.db.Save(entity).Error
	return
}

func (repo *confirmTokenPostgreRepository) DeleteToken(entity *ConfirmTokenEntity) (err error,) {
	err = repo.db.Delete(entity).Error
	return
}

func (repo *confirmTokenPostgreRepository) GetToken(tokenString string) (
	entity *ConfirmTokenEntity, err error,
) {
	err = repo.db.
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}

func (repo *confirmTokenPostgreRepository) GetTokenWithUser(tokenString string) (
	entity *ConfirmTokenEntity, err error,
) {
	err = repo.db.
		Joins("User").
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}


// Forgot Password ====================================================================================

type ForgotPasswordTokenEntity struct {
	Token string `gorm:"primaryKey;index:,type:hash"`

	TenantId *string              `gorm:"size:36;not null"`
	Tenant   tenant.TenantEntity `gorm:"foreignKey:TenantId;not null"`

	UserId uint                    `gorm:"not null"`
	User   user.TenantMemberEntity `gorm:"foreignKey:UserId;not null"`

	CreatedAt time.Time
	ExpiresAt time.Time
}


func newForgotPasswordTokenEntityFromDomain(tokenObj ForgotPasswordToken) (ForgotPasswordTokenEntity) {
	var tenantId *string
	if tokenObj.tenantId != nil {
		str := tokenObj.tenantId.String()
		tenantId = &str
	}
	return ForgotPasswordTokenEntity{
		Token: tokenObj.hashedToken,
		TenantId: tenantId,
		UserId: tokenObj.userId,
		ExpiresAt: tokenObj.expiryDate,
	}
}

func (entity *ForgotPasswordTokenEntity) ToConfirmToken() (ForgotPasswordToken, error) {
	tenantId, err := uuid.Parse(*entity.TenantId)
	if err != nil {
		return ForgotPasswordToken{}, err
	}
	return ForgotPasswordToken{
		hashedToken: entity.Token,
		tenantId: &tenantId,
		userId: entity.UserId,
		expiryDate: entity.ExpiresAt,
	}, nil
}

func (ForgotPasswordTokenEntity) TableName() string { return "forgot_password_tokens" }

// repository -----------------------------------------------------------------------------------------
type passwordTokenPostgreRepository struct {
	db *gorm.DB
}

func newPasswordTokenPostgreRepository(db *gorm.DB) *passwordTokenPostgreRepository {
	return &passwordTokenPostgreRepository{
		db: db,
	}
}

func (repo *passwordTokenPostgreRepository) SaveToken(entity *ForgotPasswordTokenEntity) (err error,) {
	err = repo.db.Save(entity).Error
	return
}

func (repo *passwordTokenPostgreRepository) DeleteToken(entity *ForgotPasswordTokenEntity) (err error,) {
	err = repo.db.Delete(entity).Error
	return
}

func (repo *passwordTokenPostgreRepository) GetToken(tokenString string) (
	entity *ForgotPasswordTokenEntity, err error,
) {
	err = repo.db.
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}

func (repo *passwordTokenPostgreRepository) GetTokenWithUser(tokenString string) (
	entity *ForgotPasswordTokenEntity, err error,
) {
	err = repo.db.
		Joins("User").
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}

