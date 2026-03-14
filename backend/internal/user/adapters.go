package user

import (
	"github.com/google/uuid"
)

type CreateUserPort interface {
	CreateUser(user User) (*User, error)
}

type DeleteUserPort interface {
	DeleteUser(userId int) (*User, error)
}

type GetUserPort interface {
	GetUserById(userId int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUsers(page, limit int) ([]User, int, error)
	GetUsersByRole(role UserRole, page, limit int) ([]User, int, error)
	GetUsersByTenantId(tenantId uuid.UUID) ([]User, int, error)
}

type UserPostgreAdapter struct {
	repository *userPostgreRepository
}

func NewUserPostgreAdapter(repository *userPostgreRepository) (CreateUserPort, DeleteUserPort, GetUserPort) {
	adapter := &UserPostgreAdapter{
		repository: repository,
	}
	return adapter, adapter, adapter
}

func (adapter *UserPostgreAdapter) CreateUser(user User) (*User, error) {
	return nil, nil
}

func (adapter *UserPostgreAdapter) DeleteUser(userId int) (*User, error) {
	return nil, nil
}

func (adapter *UserPostgreAdapter) GetUserByEmail(email string) (*User, error) {
	return nil, nil
}

func (adapter *UserPostgreAdapter) GetUserById(userId int) (*User, error) {
	return nil, nil
}

func (adapter *UserPostgreAdapter) GetUsers(page, limit int) ([]User, int, error) {
	return nil, 0, nil
}

func (adapter *UserPostgreAdapter) GetUsersByRole(role UserRole, page, limit int) ([]User, int, error) {
	return nil, 0, nil
}

func (adapter *UserPostgreAdapter)  GetUsersByTenantId(tenantId uuid.UUID) ([]User, int, error) {
	return nil, 0, nil
}


// Compile-time checks
var (
	_ CreateUserPort = (*UserPostgreAdapter)(nil)
	_ DeleteUserPort = (*UserPostgreAdapter)(nil)
	_ GetUserPort    = (*UserPostgreAdapter)(nil)
)
