package auth

import (
	"time"

	"backend/internal/tenant"
	"backend/internal/user"

	clouddb "backend/internal/infra/database/cloud_db/connection"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

/*
Entity per token di conferma account all'interno di un tenant
*/
type TenantConfirmTokenEntity struct {
	Token string `gorm:"primaryKey;index:,type:hash"`

	TenantId *string             `gorm:"size:36"`
	Tenant   tenant.TenantEntity `gorm:"foreignKey:TenantId;not null"`

	UserId       uint                    `gorm:"not null"`
	TenantMember user.TenantMemberEntity `gorm:"foreignKey:UserId;not null"`

	CreatedAt time.Time
	ExpiresAt time.Time
}

func ConfirmAccountTokenToTenantEntity(tokenObj ConfirmAccountToken) *TenantConfirmTokenEntity {
	var tenantId *string
	if tokenObj.TenantId != nil {
		str := tokenObj.TenantId.String()
		tenantId = &str
	}
	return &TenantConfirmTokenEntity{
		Token:     tokenObj.HashedToken,
		TenantId:  tenantId,
		UserId:    tokenObj.UserId,
		ExpiresAt: tokenObj.ExpiryDate,
	}
}

func TenantConfirmTokenEntityToConfirmAccountToken(entity *TenantConfirmTokenEntity) (
	token ConfirmAccountToken, err error,
) {
	tenantId, err := uuid.Parse(*entity.TenantId)
	if err != nil {
		return ConfirmAccountToken{}, err
	}
	return ConfirmAccountToken{
		HashedToken: entity.Token,
		TenantId:    &tenantId,
		UserId:      entity.UserId,
		ExpiryDate:  entity.ExpiresAt,
	}, nil
}

func (TenantConfirmTokenEntity) TableName() string { return "confirm_tokens" }

// repository -----------------------------------------------------------------------------------------

type tenantConfirmTokenPgRepository struct {
	db clouddb.CloudDBConnection
}
var _ TenantConfirmTokenRepository = (*tenantConfirmTokenPgRepository)(nil)

func newTenantConfirmTokenPgRepository(db clouddb.CloudDBConnection) *tenantConfirmTokenPgRepository {
	return &tenantConfirmTokenPgRepository{
		db: db,
	}
}

func (repo *tenantConfirmTokenPgRepository) SaveToken(entity *TenantConfirmTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(*entity.TenantId, &TenantConfirmTokenEntity{})).
		Save(entity).
		Error
	return
}

func (repo *tenantConfirmTokenPgRepository) DeleteToken(entity *TenantConfirmTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(*entity.TenantId, &TenantConfirmTokenEntity{})).
		Delete(entity).
		Error
	return
}

func (repo *tenantConfirmTokenPgRepository) GetToken(tenantId string, tokenString string) (
	entity *TenantConfirmTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantConfirmTokenEntity{})).
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}

func (repo *tenantConfirmTokenPgRepository) GetTokenWithUser(tenantId string, tokenString string) (
	entity *TenantConfirmTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Joins("User").
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantConfirmTokenEntity{})).
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}

// Forgot Password ====================================================================================

/*
Entity per i token di cambio password dimenticata all'interno di un tenant
*/
type TenantPasswordTokenEntity struct {
	Token string `gorm:"primaryKey;index:,type:hash"`

	TenantId *string             `gorm:"size:36;not null"`
	Tenant   tenant.TenantEntity `gorm:"foreignKey:TenantId;not null"`

	UserId       uint                    `gorm:"not null"`
	TenantMember user.TenantMemberEntity `gorm:"foreignKey:UserId;not null"`

	CreatedAt time.Time
	ExpiresAt time.Time
}

func ForgotPasswordTokenToTenantTokenEntity(tokenObj ForgotPasswordToken) *TenantPasswordTokenEntity {
	var tenantId *string
	if tokenObj.TenantId != nil {
		str := tokenObj.TenantId.String()
		tenantId = &str
	}
	return &TenantPasswordTokenEntity{
		Token:     tokenObj.HashedToken,
		TenantId:  tenantId,
		UserId:    tokenObj.UserId,
		ExpiresAt: tokenObj.ExpiryDate,
	}
}

func TenantPasswordTokenEntityToForgotPasswordToken(entity *TenantPasswordTokenEntity) (ForgotPasswordToken, error) {
	tenantId, err := uuid.Parse(*entity.TenantId)
	if err != nil {
		return ForgotPasswordToken{}, err
	}
	return ForgotPasswordToken{
		HashedToken: entity.Token,
		TenantId:    &tenantId,
		UserId:      entity.UserId,
		ExpiryDate:  entity.ExpiresAt,
	}, nil
}

func (TenantPasswordTokenEntity) TableName() string { return "forgot_password_tokens" }

// repository -----------------------------------------------------------------------------------------
type tenantPasswordTokenPgRepository struct {
	db clouddb.CloudDBConnection
}
var _ TenantPasswordTokenRepository = (*tenantPasswordTokenPgRepository)(nil)

func newTenantPasswordTokenPgRepository(db clouddb.CloudDBConnection) *tenantPasswordTokenPgRepository {
	return &tenantPasswordTokenPgRepository{
		db: db,
	}
}

func (repo *tenantPasswordTokenPgRepository) SaveToken(entity *TenantPasswordTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(*entity.TenantId, &TenantPasswordTokenEntity{})).
		Save(entity).
		Error
	return
}

func (repo *tenantPasswordTokenPgRepository) DeleteToken(entity *TenantPasswordTokenEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(*entity.TenantId, &TenantPasswordTokenEntity{})).
		Delete(entity).
		Error
	return
}

func (repo *tenantPasswordTokenPgRepository) GetToken(tenantId string, tokenString string) (
	entity *TenantPasswordTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantPasswordTokenEntity{})).
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}

func (repo *tenantPasswordTokenPgRepository) GetTokenWithUser(tenantId string, tokenString string) (
	entity *TenantPasswordTokenEntity, err error,
) {
	db := (*gorm.DB)(repo.db)
	err = db.
		Scopes(clouddb.WithTenantSchema(tenantId, &TenantPasswordTokenEntity{})).
		Joins("User").
		Where("token = ?", tokenString).
		First(&entity).
		Error
	return
}
