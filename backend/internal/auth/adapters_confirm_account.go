package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/shared/identity"
	"backend/internal/user"

	"github.com/google/uuid"
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
	rawToken, hashedTokenString, err := adapter.tokenGenerator.GenerateToken()
	if err != nil {
		return "", err
	}

	// 2. Save token
	switch user.Role {
	case identity.ROLE_SUPER_ADMIN:
		entity := SuperAdminConfirmTokenEntity{
			Token:  hashedTokenString,
			UserId: user.Id,
		}
		err = adapter.superAdminRepo.SaveToken(&entity)
		if err != nil {
			return "", err
		}

	case identity.ROLE_TENANT_ADMIN, identity.ROLE_TENANT_USER:
		tenantIdString := user.TenantId.String()
		entity := TenantConfirmTokenEntity{
			Token:    hashedTokenString,
			TenantId: &tenantIdString,
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
	// 1. Hash token
	hashedTokenString, err := adapter.hasher.HashSecret(tokenString)
	if err != nil {
		return
	}

	// 2. Get token
	tokenEntity, err := adapter.tenantMemberRepo.GetTokenWithUser(tenantId.String(), hashedTokenString)
	if err != nil {
		return
	}

	// 3. Get user from token
	userFound, err = user.TenantMemberEntityToUser(&tokenEntity.TenantMember)
	return
}

func (adapter *ConfirmTokenPgAdapter) GetSuperAdminByConfirmAccountToken(tokenString string) (
	userFound user.User, err error,
) {
	// 1. Hash token
	hashedTokenString, err := adapter.hasher.HashSecret(tokenString)
	if err != nil {
		return
	}

	// 2. Get token
	tokenEntity, err := adapter.superAdminRepo.GetTokenWithUser(hashedTokenString)
	if err != nil {
		return
	}

	// 3. Get user from token
	userFound, err = user.SuperAdminEntityToUser(&tokenEntity.SuperAdmin)
	return
}

// Get token ------------------------------------------------------------------------------------------

func (adapter *ConfirmTokenPgAdapter) GetTenantConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
	token ConfirmAccountToken, err error,
) {
	// 1. Hash token
	hashedTokenString, err := adapter.hasher.HashSecret(tokenString)
	if err != nil {
		return
	}

	// 2. Get token
	tokenEntity, err := adapter.tenantMemberRepo.GetToken(tenantId.String(), hashedTokenString)
	if err != nil {
		return
	}
	token, err = TenantConfirmTokenEntityToConfirmAccountToken(tokenEntity)
	return
}

func (adapter *ConfirmTokenPgAdapter) GetSuperAdminConfirmAccountToken(tokenString string) (
	token ConfirmAccountToken, err error,
) {
	// 1. Hash token
	hashedTokenString, err := adapter.hasher.HashSecret(tokenString)
	if err != nil {
		return
	}

	// 2. Get token
	tokenEntity, err := adapter.superAdminRepo.GetToken(hashedTokenString)
	if err != nil {
		return
	}
	token = SuperAdminConfirmTokenEntityToConfirmAccountToken(tokenEntity)
	return
}
