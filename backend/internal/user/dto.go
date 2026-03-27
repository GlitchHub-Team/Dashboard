package user

import (
	"backend/internal/common/dto"
)

// Request DTO ========================================================================================

// Create --------------------------------------------------------------------------
type CreateTenantUserDTO struct {
	dto.EmailField
	dto.UsernameField
	dto.TenantIdField
}

type CreateTenantAdminDTO struct {
	dto.EmailField
	dto.UsernameField
	dto.TenantIdField
}

type CreateSuperAdminDTO struct {
	dto.EmailField
	dto.UsernameField
}

// Delete --------------------------------------------------------------------------

type DeleteTenantUserDTO struct {
	dto.TenantIdField
	dto.UserIdField
}

type DeleteTenantAdminDTO struct {
	dto.TenantIdField
	dto.UserIdField
}

type DeleteSuperAdminDTO struct {
	dto.UserIdField
}

// Get ------------------------------------------------------------------------------
type GetUserByIdDTO struct {
	dto.UserIdField
}

type GetTenantUserDTO struct {
	dto.TenantIdField
	dto.UserIdField
}

type GetTenantAdminDTO struct {
	dto.TenantIdField
	dto.UserIdField
}

type GetSuperAdminDTO struct {
	dto.UserIdField
}

type GetUsersDTO struct {
	dto.Pagination
	dto.UserRoleField
}

type GetUsersByTenantIdDTO struct {
	dto.Pagination
	dto.TenantIdField
}

// Response DTO
type UserResponseDTO struct {
	dto.UserIdField
	dto.EmailField
	dto.UserRoleField
	dto.TenantIdField
}
func NewUserResponseDTO(user User) UserResponseDTO {
	return UserResponseDTO{
		dto.UserIdField{UserId: user.Id},
		dto.EmailField{Email: user.Email},
		dto.UserRoleField{UserRole: string(user.Role)},
		dto.TenantIdField{TenantId: user.TenantId.String()},
	}
}

type UserListResponseDTO struct{
	dto.ListInfo
	Users []UserResponseDTO
}
func NewUserListResponseDTO(userList []User, total int) UserListResponseDTO {
	var userDtos []UserResponseDTO

	for _, user := range userList {
		userDtos = append(userDtos, NewUserResponseDTO(user))
	}

	return UserListResponseDTO{
		Users: userDtos,
		ListInfo: dto.ListInfo{
			Count: len(userList),
			Total: total,
		},
	}
}