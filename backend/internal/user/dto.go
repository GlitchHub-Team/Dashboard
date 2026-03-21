package user

import (
	"backend/internal/common/dto"
	"backend/internal/identity"
)

// Request DTO ========================================================================================

// Create ---------------------------------------------------------------------------------------------

type CreateUserBodyDTO struct {
	dto.EmailField
	dto.UsernameField
}

// type CreateUserDTO struct {
// 	dto.EmailField
// 	dto.UsernameField
// }

// Delete ---------------------------------------------------------------------------------------------

// type DeleteTenantUserDTO struct {
// 	dto.TenantIdField
// 	dto.UserIdField
// }

// type DeleteTenantAdminDTO struct {
// 	dto.TenantIdField
// 	dto.UserIdField
// }

// type DeleteSuperAdminDTO struct {
// 	dto.UserIdField
// }

// Get single ---------------------------------------------------------------------------------------------
// type GetTenantUserDTO struct {
// 	dto.TenantIdField
// 	dto.UserIdField
// }

// type GetTenantAdminDTO struct {
// 	dto.TenantIdField
// 	dto.UserIdField
// }

// type GetSuperAdminDTO struct {
// 	dto.UserIdField
// }

// Get multiple ---------------------------------------------------------------------------------------
// type GetTenantUsersByTenantQueryDTO struct {
// 	dto.Pagination
// }

// type GetTenantAdminsByTenantQueryDTO struct {
// 	dto.Pagination
// }

// type GetSuperAdminListQueryDTO struct {
// 	dto.Pagination
// }

type GetUserListQueryDTO struct {
	dto.Pagination
}

// Response DTO
type UserResponseDTO struct {
	dto.UserIdField
	dto.EmailField
	dto.UsernameField
	dto.UserRoleField
	dto.TenantIdField
}

func NewUserResponseDTO(user User) UserResponseDTO {
	response := UserResponseDTO{
		UserIdField: dto.UserIdField{UserId: user.Id},
		UsernameField: dto.UsernameField{Username: user.Name},
		EmailField: dto.EmailField{Email: user.Email},
		UserRoleField: dto.UserRoleField{UserRole: string(user.Role)},
	}
	if user.Role != identity.ROLE_SUPER_ADMIN {
		response.TenantIdField = dto.TenantIdField{
			TenantId: user.TenantId.String(),
		}
	}

	return response
}

type UserListResponseDTO struct {
	dto.ListInfo
	Users []UserResponseDTO `json:"users" binding:"required,min=0"`
}

func NewUserListResponseDTO(userList []User, total uint) UserListResponseDTO {
	userDtos := make([]UserResponseDTO, 0) // Importante creare un empty slice e non un nil slice!

	for _, user := range userList {
		userDtos = append(userDtos, NewUserResponseDTO(user))
	}

	return UserListResponseDTO{
		Users: userDtos,
		ListInfo: dto.ListInfo{
			Count: uint(len(userList)),
			Total: total,
		},
	}
}

