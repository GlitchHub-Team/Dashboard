package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/user"
	// "github.com/google/uuid"
)

// ConfirmToken =======================================================================================

type ConfirmTokenPostgreAdapter struct {
	tokenGenerator crypto.SecurityTokenGenerator
	repository     *confirmTokenPostgreRepository
}

var _ ConfirmAccountTokenPort = (*ConfirmTokenPostgreAdapter)(nil) // Compile-time check

func NewConfirmAccountTokenPostgreAdapter(
	tokenGenerator crypto.SecurityTokenGenerator,
	repository *confirmTokenPostgreRepository,
) *ConfirmTokenPostgreAdapter {
	return &ConfirmTokenPostgreAdapter{
		tokenGenerator: tokenGenerator,
		repository:     repository,
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
	tenantIdString := user.TenantId.String()
	entity := ConfirmTokenEntity{
		Token:    hashedTokenString,
		TenantId: &tenantIdString,
		UserId:   user.Id,
	}
	err = adapter.repository.SaveToken(&entity)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func (adapter *ConfirmTokenPostgreAdapter) DeleteConfirmAccountToken(token ConfirmAccountToken) (err error) {
	// Delete token
	entity := newConfirmTokenEntityFromDomain(token)
	err = adapter.repository.DeleteToken(&entity)
	return err
}

func (adapter *ConfirmTokenPostgreAdapter) GetUserByConfirmAccountToken(tokenString string) (
	userFound user.User, err error,
) {
	// 1. Get token
	tokenEntity, err := adapter.repository.GetTokenWithUser(tokenString)
	if err != nil {
		return user.User{}, err
	}

	// 2. Get user from token
	userFound, err = tokenEntity.User.ToUser()
	return
}

func (adapter *ConfirmTokenPostgreAdapter) GetConfirmAccountToken(tokenString string) (
	token ConfirmAccountToken, err error,
) {
	tokenEntity, err := adapter.repository.GetToken(tokenString)
	if err != nil {
		return ConfirmAccountToken{}, err
	}
	token, err = tokenEntity.ToConfirmToken()
	return
}

// ChangePasswordToken ============================================================================

type ChangePasswordTokenPostgreAdapter struct {
	tokenGenerator crypto.SecurityTokenGenerator

	repository *passwordTokenPostgreRepository
}

var _ ChangePasswordTokenPort = (*ChangePasswordTokenPostgreAdapter)(nil) // Compile-time check

func NewChangePasswordTokenPostgreAdapter(tokenGenerator crypto.SecurityTokenGenerator, repository *passwordTokenPostgreRepository) *ChangePasswordTokenPostgreAdapter {
	return &ChangePasswordTokenPostgreAdapter{
		tokenGenerator: tokenGenerator,
		repository:     repository,
	}
}

func (adapter *ChangePasswordTokenPostgreAdapter) NewChangePasswordToken(user user.User) (
	rawToken string, err error,
) {
	// 1. Generate token
	rawToken, hashedTokenString, err := adapter.tokenGenerator.GenerateToken()
	if err != nil {
		return "", err
	}

	// 2. Save token
	tenantIdString := user.TenantId.String()
	entity := ForgotPasswordTokenEntity{
		Token:    hashedTokenString,
		TenantId: &tenantIdString,
		UserId:   user.Id,
	}
	err = adapter.repository.SaveToken(&entity)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) DeleteChangePasswordToken(tokenObj ForgotPasswordToken) (err error) {
	// Delete token
	entity := newForgotPasswordTokenEntityFromDomain(tokenObj)
	err = adapter.repository.DeleteToken(&entity)
	return
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetUserByChangePasswordToken(tokenString string) (
	userFound user.User, err error,
) {
	// 1. Get token
	tokenEntity, err := adapter.repository.GetTokenWithUser(tokenString)
	if err != nil {
		return user.User{}, err
	}

	// 2. Get user from token
	userFound, err = tokenEntity.User.ToUser()
	return
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetChangePasswordToken(tokenString string) (
	token ForgotPasswordToken, err error,
) {
	tokenEntity, err := adapter.repository.GetToken(tokenString)
	if err != nil {
		return ForgotPasswordToken{}, err
	}
	token, err = tokenEntity.ToConfirmToken()
	return
}
