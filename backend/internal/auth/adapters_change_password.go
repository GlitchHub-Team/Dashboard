package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../../tests/auth/mocks/repository_change_password.go -package=mocks . SuperAdminPasswordTokenRepository,TenantPasswordTokenRepository

// Interfacce =========================================================================================

type TenantPasswordTokenRepository interface {
	SaveToken(entity *TenantPasswordTokenEntity) (err error)
	DeleteToken(entity *TenantPasswordTokenEntity) (err error)
	GetToken(tenantId string, tokenString string) (
		entity *TenantPasswordTokenEntity, err error,
	)
	GetTokenWithUser(tenantId string, tokenString string) (
		entity *TenantPasswordTokenEntity, err error,
	)
}

type SuperAdminPasswordTokenRepository interface {
	SaveToken(entity *SuperAdminPasswordTokenEntity) (err error)
	DeleteToken(entity *SuperAdminPasswordTokenEntity) (err error)
	GetToken(tokenString string) (
		entity *SuperAdminPasswordTokenEntity, err error,
	)
	GetTokenWithUser(tokenString string) (
		entity *SuperAdminPasswordTokenEntity, err error,
	)
}

// Adapter ============================================================================================

type ChangePasswordTokenPgAdapter struct {
	hasher         crypto.SecretHasher
	tokenGenerator crypto.SecurityTokenGenerator

	// repository *superAdminPasswordTokenPgRepository
	tenantMemberRepo TenantPasswordTokenRepository
	superAdminRepo   SuperAdminPasswordTokenRepository
}

var _ ForgotPasswordTokenPort = (*ChangePasswordTokenPgAdapter)(nil) // Compile-time check

func NewChangePasswordTokenPgAdapter(
	hasher crypto.SecretHasher,
	tokenGenerator crypto.SecurityTokenGenerator,
	tenantMemberRepo TenantPasswordTokenRepository,
	superAdminRepo SuperAdminPasswordTokenRepository,
) *ChangePasswordTokenPgAdapter {
	return &ChangePasswordTokenPgAdapter{
		hasher:           hasher,
		tokenGenerator:   tokenGenerator,
		tenantMemberRepo: tenantMemberRepo,
		superAdminRepo:   superAdminRepo,
	}
}

func (adapter *ChangePasswordTokenPgAdapter) NewForgotPasswordToken(user user.User) (
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
		entity := SuperAdminPasswordTokenEntity{
			Token:  rawToken,
			UserId: user.Id,
		}
		err = adapter.superAdminRepo.SaveToken(&entity)
		if err != nil {
			return "", err
		}

	case identity.ROLE_TENANT_ADMIN, identity.ROLE_TENANT_USER:
		tenantIdString := user.TenantId.String()
		entity := TenantPasswordTokenEntity{
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

func (adapter *ChangePasswordTokenPgAdapter) DeleteForgotPasswordToken(token ForgotPasswordToken) (err error) {
	// Super Admin
	if token.TenantId == nil {
		entity := ForgotPasswordTokenToSuperAdminTokenEntity(token)
		err = adapter.superAdminRepo.DeleteToken(entity)
	} else
	// Tenant Member
	{
		entity := ForgotPasswordTokenToTenantTokenEntity(token)
		err = adapter.tenantMemberRepo.DeleteToken(entity)
	}
	return err
}

// Get user -------------------------------------------------------------------------------------------

func (adapter *ChangePasswordTokenPgAdapter) GetTenantMemberByForgotPasswordToken(tenantId uuid.UUID, tokenString string) (
	userFound user.User, err error,
) {
	tokenEntity, err := adapter.tenantMemberRepo.GetTokenWithUser(tenantId.String(), tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || tokenEntity.Token == "" {
			return user.User{}, ErrTokenNotFound
		}
		return
	}
	if tokenEntity.Token == "" {
		return user.User{}, ErrTokenNotFound
	}
	tokenEntity.TenantId = tenantId.String()

	userFound, err = user.TenantMemberEntityToUser(&tokenEntity.TenantMember)
	return
}

func (adapter *ChangePasswordTokenPgAdapter) GetSuperAdminByForgotPasswordToken(tokenString string) (
	userFound user.User, err error,
) {
	tokenEntity, err := adapter.superAdminRepo.GetTokenWithUser(tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || tokenEntity.Token == "" {
			return user.User{}, ErrTokenNotFound
		}
		return
	}
	if tokenEntity.Token == "" {
		return user.User{}, ErrTokenNotFound
	}

	// 3. Get user from token
	userFound, _ = user.SuperAdminEntityToUser(&tokenEntity.SuperAdmin)
	return
}

// Get token ------------------------------------------------------------------------------------------

func (adapter *ChangePasswordTokenPgAdapter) GetTenantForgotPasswordToken(tenantId uuid.UUID, tokenString string) (
	token ForgotPasswordToken, err error,
) {
	tokenEntity, err := adapter.tenantMemberRepo.GetToken(tenantId.String(), tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound){
			return ForgotPasswordToken{}, ErrTokenNotFound
		}
		return ForgotPasswordToken{}, err
	}
	if tokenEntity.Token == "" {
		return ForgotPasswordToken{}, ErrTokenNotFound
	}
	tokenEntity.TenantId = tenantId.String()

	token, err = TenantPasswordTokenEntityToForgotPasswordToken(tokenEntity)
	return
}

func (adapter *ChangePasswordTokenPgAdapter) GetSuperAdminForgotPasswordToken(tokenString string) (
	token ForgotPasswordToken, err error,
) {
	tokenEntity, err := adapter.superAdminRepo.GetToken(tokenString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ForgotPasswordToken{}, ErrTokenNotFound
		}
		return
	}
	if tokenEntity.Token == "" {
		return ForgotPasswordToken{}, ErrTokenNotFound
	}
	token = SuperAdminPasswordTokenEntityToForgotPasswordToken(tokenEntity)
	return
}
