package user

import (
	"backend/internal/common/dto"
)

// Request DTO
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

type DeleteUserDTO struct {
	dto.UserIdField
}

type GetUserByIdDTO struct {
	dto.UserIdField
}

type GetUsersDTO struct {
	dto.Pagination
	dto.UserRoleField
}

type GetUsersByTenantIdDTO struct {
	dto.Pagination
	dto.UserIdField
}

// Response DTO
type UserResponseDTO struct {
	dto.UserIdField
	dto.EmailField
	dto.UserRoleField
	dto.TenantIdField
}
func NewUserResponseDTO(user *User) UserResponseDTO {
	return UserResponseDTO{
		dto.UserIdField{UserId: user.id},
		dto.EmailField{Email: user.email},
		dto.UserRoleField{UserRole: string(user.role)},
		dto.TenantIdField{TenantId: user.tenantId.String()},
	}
}

type UserListResponseDTO struct{
	dto.ListInfo
	Users []UserResponseDTO
}
func NewUserListResponseDTO(userList []User, total int) UserListResponseDTO {
	var userDtos []UserResponseDTO

	for _, user := range userList {
		userDtos = append(userDtos, NewUserResponseDTO(&user))
	}

	return UserListResponseDTO{
		Users: userDtos,
		ListInfo: dto.ListInfo{
			Count: len(userList),
			Total: total,
		},
	}
}