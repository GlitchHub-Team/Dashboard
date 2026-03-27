package auth

//go:generate mockgen -destination=../../tests/auth/mocks/ports.go -package=mocks . ConfirmTokenPort,ChangePasswordTokenPort

// ConfirmToken ============================================================================
type ConfirmTokenPort interface {
	NewConfirmAccountToken(userId uint) (string, error)
	DeleteConfirmAccountToken(tokenId int) error
	GetConfirmAccountTokenByUserId(userId uint) (*ConfirmToken, error)
}

type ConfirmTokenPostgreAdapter struct {
	repository *confirmTokenPostgreRepository
}

func NewConfirmAccountTokenPostgreAdapter(
	repository *confirmTokenPostgreRepository,
) ConfirmTokenPort {
	return &ConfirmTokenPostgreAdapter{
		repository: repository,
	}
}

func (adapter *ConfirmTokenPostgreAdapter) NewConfirmAccountToken(userId uint) (string, error) {
	return "", nil
}

func (adapter *ConfirmTokenPostgreAdapter) DeleteConfirmAccountToken(tokenId int) error {
	return nil
}

func (adapter *ConfirmTokenPostgreAdapter) GetConfirmAccountTokenByUserId(userId uint) (*ConfirmToken, error) {
	return nil, nil
}

// Compile-time checks
var _ ConfirmTokenPort = (*ConfirmTokenPostgreAdapter)(nil)

// ChangePasswordToken ============================================================================
type ChangePasswordTokenPort interface {
	SaveChangePasswordToken(token ChangePasswordToken) (*ChangePasswordToken, error)
	DeleteChangePasswordToken(tokenId int) error
	GetChangePasswordTokenByUserId(userId uint) (*ChangePasswordToken, error)
}

type ChangePasswordTokenPostgreAdapter struct {
	repository passwordTokenPostgreRepository
}

func (adapter *ChangePasswordTokenPostgreAdapter) SaveChangePasswordToken(token ChangePasswordToken) (*ChangePasswordToken, error) {
	return nil, nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) DeleteChangePasswordToken(tokenId int) error {
	return nil
}

func (adapter *ChangePasswordTokenPostgreAdapter) GetChangePasswordTokenByUserId(userId uint) (*ChangePasswordToken, error) {
	return nil, nil
}

// Compile-time checks
var _ ChangePasswordTokenPort = (*ChangePasswordTokenPostgreAdapter)(nil)
