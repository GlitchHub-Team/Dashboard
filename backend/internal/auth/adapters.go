package auth

import (
	"backend/internal/user"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/auth/mocks/ports.go -package=mocks . ConfirmAccountTokenPort,ChangePasswordTokenPort

// ConfirmToken ============================================================================

type ConfirmTokenPostgreAdapter struct {
	repository *confirmTokenPostgreRepository
}

func NewConfirmAccountTokenPostgreAdapter(
	repository *confirmTokenPostgreRepository,
) *ConfirmTokenPostgreAdapter {
	return &ConfirmTokenPostgreAdapter{
		repository: repository,
	}
}

func (adapter *ConfirmTokenPostgreAdapter) NewConfirmAccountToken(tenantId *uuid.UUID, userId uint) (string, error) {
	return "", nil
}

func (adapter *ConfirmTokenPostgreAdapter) DeleteConfirmAccountToken(token string) error {
	return nil
}

func (adapter *ConfirmTokenPostgreAdapter) GetConfirmAccountTokenByUser(tenantId *uuid.UUID, userId uint) (
	ConfirmAccountToken, error,
) {
	return ConfirmAccountToken{}, nil
}

func (adapter *ConfirmTokenPostgreAdapter) GetUserByConfirmAccountToken(token string) (user.User, error) {
	return user.User{}, nil
}

func (adapter *ConfirmTokenPostgreAdapter) GetConfirmAccountToken(token string) (ConfirmAccountToken, error) {
	return ConfirmAccountToken{}, nil
}

// Compile-time checks
var _ ConfirmAccountTokenPort = (*ConfirmTokenPostgreAdapter)(nil)

// ChangePasswordToken ============================================================================

type ChangePasswordTokenPostgreAdapter struct {
	repository passwordTokenPostgreRepository
}

func (adapter *ChangePasswordTokenPostgreAdapter) SaveChangePasswordToken(token ForgotPasswordToken) (ForgotPasswordToken, error) {
	return ForgotPasswordToken{}, nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) DeleteChangePasswordToken(token ForgotPasswordToken) error {
	return nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetChangePasswordTokenByUser(tenantId *uuid.UUID, userId uint) (ForgotPasswordToken, error) {
	return ForgotPasswordToken{}, nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetChangePasswordToken(hashedTokenString string) (ForgotPasswordToken, error) {
	return ForgotPasswordToken{}, nil
}

// Compile-time checks
var _ ChangePasswordTokenPort = (*ChangePasswordTokenPostgreAdapter)(nil)
