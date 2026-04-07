package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../../tests/auth/mocks/repository_confirm_account.go -package=mocks . SuperAdminConfirmTokenRepository,TenantConfirmTokenRepository

// Interfacce =========================================================================================

type TenantConfirmTokenRepository interface {
	SaveToken(entity *TenantConfirmTokenEntity) (err error)
	DeleteToken(entity *TenantConfirmTokenEntity) (err error)
	GetToken(tenantId string, tokenString string) (
		entity *TenantConfirmTokenEntity, err error,
	)
	GetTokenWithUser(tenantId string, tokenString string) (
		entity *TenantConfirmTokenEntity, err error,
	)
}

type SuperAdminConfirmTokenRepository interface {
	SaveToken(entity *SuperAdminConfirmTokenEntity) (err error)
	DeleteToken(entity *SuperAdminConfirmTokenEntity) (err error)
	GetToken(tokenString string) (
		entity *SuperAdminConfirmTokenEntity, err error,
	)
	GetTokenWithUser(tokenString string) (
		entity *SuperAdminConfirmTokenEntity, err error,
	)
}

// Adapter ============================================================================================

type ConfirmTokenPgAdapter struct {
	hasher           crypto.SecretHasher
	tokenGenerator   crypto.SecurityTokenGenerator
	tenantMemberRepo TenantConfirmTokenRepository
	superAdminRepo   SuperAdminConfirmTokenRepository
}

var _ ConfirmAccountTokenPort = (*ConfirmTokenPgAdapter)(nil) // Compile-time check

func NewConfirmAccountTokenPgAdapter(
	hasher crypto.SecretHasher,
	tokenGenerator crypto.SecurityTokenGenerator,
	tenantMemberRepo TenantConfirmTokenRepository,
	superAdminRepo SuperAdminConfirmTokenRepository,
) *ConfirmTokenPgAdapter {
	return &ConfirmTokenPgAdapter{
		hasher:           hasher,
		tokenGenerator:   tokenGenerator,
		tenantMemberRepo: tenantMemberRepo,
		superAdminRepo:   superAdminRepo,
	}
}

func (adapter *ConfirmTokenPgAdapter) NewConfirmAccountToken(user user.User) (
	rawToken string, err error,
) {
	// 1. Generate token
	rawToken, _, err = adapter.tokenGenerator.GenerateToken()
	if err != nil {
		return "", err
	}

	// 2. Save token
	switch user.Role {
	case identity.ROLE_SUPER_ADMIN:
		entity := SuperAdminConfirmTokenEntity{
			Token:  rawToken,
			UserId: user.Id,
		}
		err = adapter.superAdminRepo.SaveToken(&entity)
		if err != nil {
			return "", err
		}

	case identity.ROLE_TENANT_ADMIN, identity.ROLE_TENANT_USER:
		tenantIdString := user.TenantId.String()
		entity := TenantConfirmTokenEntity{
			Token:    rawToken,
			TenantId: tenantIdString,
			UserId:   user.Id,
		}
		err = adapter.tenantMemberRepo.SaveToken(&entity)
		if err != nil {
			return "", err
		}
	default:
		return "", identity.ErrUnknownRole
	}

	return rawToken, nil
}

func (adapter *ConfirmTokenPgAdapter) DeleteConfirmAccountToken(token ConfirmAccountToken) (err error) {
	// Super Admin
	if token.TenantId == nil {
		entity := ConfirmAccountTokenToSuperAdminEntity(token)
		err = adapter.superAdminRepo.DeleteToken(entity)
	} else
	// Tenant Member
	{
		entity := ConfirmAccountTokenToTenantEntity(token)
		err = adapter.tenantMemberRepo.DeleteToken(entity)
	}
	return err
}

// Get user -------------------------------------------------------------------------------------------

func (adapter *ConfirmTokenPgAdapter) GetTenantMemberByConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
	userFound user.User, err error,
) {
	tokenEntity, err := adapter.tenantMemberRepo.GetTokenWithUser(tenantId.String(), tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || tokenEntity.Token == "" {
			return user.User{}, ErrTokenNotFound
		}
		return
	}

	userFound, err = user.TenantMemberEntityToUser(&tokenEntity.TenantMember)
	return
}

func (adapter *ConfirmTokenPgAdapter) GetSuperAdminByConfirmAccountToken(tokenString string) (
	userFound user.User, err error,
) {
	tokenEntity, err := adapter.superAdminRepo.GetTokenWithUser(tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || tokenEntity.Token == "" {
			return user.User{}, ErrTokenNotFound
		}
		return
	}

	userFound, err = user.SuperAdminEntityToUser(&tokenEntity.SuperAdmin)
	return
}

// Get token ------------------------------------------------------------------------------------------

func (adapter *ConfirmTokenPgAdapter) GetTenantConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
	token ConfirmAccountToken, err error,
) {
	tokenEntity, err := adapter.tenantMemberRepo.GetToken(tenantId.String(), tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ConfirmAccountToken{}, ErrTokenNotFound
		}
		return
	}
	tokenEntity.TenantId = tenantId.String()

	token, err = TenantConfirmTokenEntityToConfirmAccountToken(tokenEntity)
	return
}

func (adapter *ConfirmTokenPgAdapter) GetSuperAdminConfirmAccountToken(tokenString string) (
	token ConfirmAccountToken, err error,
) {
	tokenEntity, err := adapter.superAdminRepo.GetToken(tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ConfirmAccountToken{}, ErrTokenNotFound
		}
		return
	}
	token = SuperAdminConfirmTokenEntityToConfirmAccountToken(tokenEntity)
	return
}
