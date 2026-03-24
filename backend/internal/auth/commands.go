package auth

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type LoginUserCommand struct {
	TenantId *uuid.UUID
	Email    string
	Password string
	Role     identity.UserRole
}

type LogoutUserCommand struct {
	identity.Requester
	TenantId uuid.UUID
	UserId   uint
}

// Confirm account ------------------------------------
type ConfirmAccountCommand struct {
	// identity.Requester
	Token       string
	NewPassword string
}

type VerifyConfirmAccountTokenCommand struct {
	// identity.Requester
	Token string
}

// Forgot password ---------------------------------------
type VerifyForgotPasswordTokenCommand struct {
	// identity.Requester
	Token string
}

type RequestForgotPasswordCommand struct {
	// identity.Requester
	TenantId *uuid.UUID
	Email    string
}

type ConfirmForgotPasswordCommand struct {
	// identity.Requester
	Token       string
	NewPassword string
}

// Change password -----------------------------------
type ChangePasswordCommand struct {
	identity.Requester
	OldPassword string
	NewPassword string
}
