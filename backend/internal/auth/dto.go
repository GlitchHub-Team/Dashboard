package auth

import (
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
)

// LOGIN/LOGOUT =======================================================================================
type LoginUserDTO struct {
	dto.TenantIdField_NotRequired
	dto.EmailField
	dto.PasswordField
}

type LogoutUserDTO struct {
	identity.Requester
}

// CONFIRM ACCOUNT ====================================================================================
type VerifyConfirmAccountTokenBodyDTO struct {
	dto.TokenFields
}

type ConfirmUserAccountBodyDTO struct {
	dto.TokenFields
	dto.NewPasswordField
}

// FORGOT PASSWORD =======================================================================================
type VerifyForgotPasswordTokenBodyDTO struct {
	dto.TokenFields
}

type RequestForgotPasswordBodyDTO struct {
	dto.TenantIdField_NotRequired
	dto.EmailField
}

type ConfirmForgotPasswordBodyDTO struct {
	dto.TokenFields
	dto.NewPasswordField
}

// CHANGE PASSWORD ====================================================================================
type ChangePasswordBodyDTO struct {
	dto.ChangePasswordFields
}

// RESPONSE DTO ==============
type LoginResponseDTO struct {
	JWT string `json:"jwt" binding:"required"`
}

type ResultDTO struct {
	Result string `json:"result" binding:"required"`
}
