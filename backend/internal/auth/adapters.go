package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/shared/identity"
	"backend/internal/user"

	"github.com/google/uuid"
	// "github.com/google/uuid"
)

// ConfirmToken =======================================================================================

type ConfirmTokenPostgreAdapter struct {
	hasher           crypto.SecretHasher
	tokenGenerator   crypto.SecurityTokenGenerator
	superAdminRepo   *superAdminConfirmTokenPgRepository
	tenantMemberRepo *tenantConfirmTokenPgRepository
}

var _ ConfirmAccountTokenPort = (*ConfirmTokenPostgreAdapter)(nil) // Compile-time check

func NewConfirmAccountTokenPostgreAdapter(
	hasher crypto.SecretHasher,
	tokenGenerator crypto.SecurityTokenGenerator,
	superAdminRepo *superAdminConfirmTokenPgRepository,
	tenantMemberRepo *tenantConfirmTokenPgRepository,
) *ConfirmTokenPostgreAdapter {
	return &ConfirmTokenPostgreAdapter{
		hasher:           hasher,
		tokenGenerator:   tokenGenerator,
		superAdminRepo:   superAdminRepo,
		tenantMemberRepo: tenantMemberRepo,
	}
}

func (adapter *ConfirmTokenPostgreAdapter) NewConfirmAccountToken(user user.User) (
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
	}

	return rawToken, nil
}

func (adapter *ConfirmTokenPostgreAdapter) DeleteConfirmAccountToken(token ConfirmAccountToken) (err error) {
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

func (adapter *ConfirmTokenPostgreAdapter) GetTenantMemberByConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
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

func (adapter *ConfirmTokenPostgreAdapter) GetSuperAdminByConfirmAccountToken(tokenString string) (
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

func (adapter *ConfirmTokenPostgreAdapter) GetTenantConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
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

func (adapter *ConfirmTokenPostgreAdapter) GetSuperAdminConfirmAccountToken(tokenString string) (
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

// ChangePasswordToken ============================================================================

type ChangePasswordTokenPostgreAdapter struct {
	hasher         crypto.SecretHasher
	tokenGenerator crypto.SecurityTokenGenerator

	// repository *superAdminPasswordTokenPgRepository
	tenantMemberRepo *tenantPasswordTokenPgRepository
	superAdminRepo   *superAdminPasswordTokenPgRepository
}

var _ ForgotPasswordTokenPort = (*ChangePasswordTokenPostgreAdapter)(nil) // Compile-time check

func NewChangePasswordTokenPostgreAdapter(
	hasher crypto.SecretHasher,
	tokenGenerator crypto.SecurityTokenGenerator,
	tenantMemberRepo *tenantPasswordTokenPgRepository,
	superAdminRepo *superAdminPasswordTokenPgRepository,
) *ChangePasswordTokenPostgreAdapter {
	return &ChangePasswordTokenPostgreAdapter{
		hasher:           hasher,
		tokenGenerator:   tokenGenerator,
		tenantMemberRepo: tenantMemberRepo,
		superAdminRepo:   superAdminRepo,
	}
}

func (adapter *ChangePasswordTokenPostgreAdapter) NewForgotPasswordToken(user user.User) (
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
		entity := SuperAdminPasswordTokenEntity{
			Token:  hashedTokenString,
			UserId: user.Id,
		}
		err = adapter.superAdminRepo.SaveToken(&entity)
		if err != nil {
			return "", err
		}

	case identity.ROLE_TENANT_ADMIN, identity.ROLE_TENANT_USER:
		tenantIdString := user.TenantId.String()
		entity := TenantPasswordTokenEntity{
			Token:    hashedTokenString,
			TenantId: &tenantIdString,
			UserId:   user.Id,
		}
		err = adapter.tenantMemberRepo.SaveToken(&entity)
		if err != nil {
			return "", err
		}
	}

	return rawToken, nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) DeleteForgotPasswordToken(token ForgotPasswordToken) (err error) {
	// Super Admin
	if token.TenantId == nil {
		entity := ForgotPasswordTokenToSuperAdminEntity(token)
		err = adapter.superAdminRepo.DeleteToken(entity)
	} else
	// Tenant Member
	{
		entity := ForgotPasswordTokenToTenantEntity(token)
		err = adapter.tenantMemberRepo.DeleteToken(entity)
	}
	return err
}

// Get user -------------------------------------------------------------------------------------------

func (adapter *ChangePasswordTokenPostgreAdapter) GetTenantMemberByForgotPasswordToken(tenantId uuid.UUID, tokenString string) (
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
		return user.User{}, err
	}

	// 3. Get user from token
	userFound, err = user.TenantMemberEntityToUser(&tokenEntity.TenantMember)
	return
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetSuperAdminByForgotPasswordToken(tokenString string) (
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

func (adapter *ChangePasswordTokenPostgreAdapter) GetTenantForgotPasswordToken(tenantId uuid.UUID, tokenString string) (
	token ForgotPasswordToken, err error,
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
	token, err = TenantPasswordTokenEntityToForgotPasswordToken(tokenEntity)
	return
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetSuperAdminForgotPasswordToken(tokenString string) (
	token ForgotPasswordToken, err error,
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
	token = SuperAdminPasswordTokenEntityToForgotPasswordToken(tokenEntity)
	return
}
