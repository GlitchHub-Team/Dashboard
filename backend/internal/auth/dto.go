package auth

import (
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
)

// LOGIN/LOGOUT =======================================================================================
type LoginUserDto struct {
	dto.TenantIdField_NotRequired
	dto.EmailField
	dto.PasswordField
	dto.UserRoleField
}

type LogoutUserDto struct {
	identity.Requester
}

// type LogoutUserDto struct {
// 	TenantId string
// 	UserId   uint
// }

// CONFIRM ACCOUNT ====================================================================================
// type VerifyConfirmAccountTokenDto struct {
// 	dto.TokenField
// }

type ConfirmUserAccountDto struct {
	dto.TokenField
	dto.NewPasswordField
}

// FORGOT PASSWORD =======================================================================================
type VerifyForgotPasswordTokenDto struct {
	dto.TokenField
}

type RequestForgotPasswordDto struct {
	dto.TenantIdField_NotRequired
	dto.EmailField
}

type ConfirmForgotPasswordDto struct {
	dto.TokenField
	dto.NewPasswordField
}

// CHANGE PASSWORD ====================================================================================
type ChangePasswordDto struct {
	dto.ChangePasswordFields
}

// RESPONSE DTO ==============
type LoginResponseDto struct {
	JWT string `json:"jwt" binding:"required"`
}

type ResultDto struct {
	Result string `json:"result" binding:"required"`
}