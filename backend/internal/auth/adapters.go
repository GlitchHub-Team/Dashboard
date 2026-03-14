package auth

// ConfirmToken ============================================================================
type ConfirmTokenPort interface {
	NewConfirmAccountToken(userId int) (string, error)
	DeleteConfirmAccountToken(tokenId int) error
	GetConfirmAccountTokenByUserId(userId int) (*ConfirmToken, error)
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

func (adapter *ConfirmTokenPostgreAdapter) NewConfirmAccountToken(userId int) (string, error) {
	return "", nil
}

func (adapter *ConfirmTokenPostgreAdapter) DeleteConfirmAccountToken(tokenId int) error {
	return nil
}

func (adapter *ConfirmTokenPostgreAdapter) GetConfirmAccountTokenByUserId(userId int) (*ConfirmToken, error) {
	return nil, nil
}

// Compile-time checks
var _ ConfirmTokenPort = (*ConfirmTokenPostgreAdapter)(nil)

// ChangePasswordToken ============================================================================
type ChangePasswordTokenPort interface {
	SaveChangePasswordToken(token ChangePasswordToken) (*ChangePasswordToken, error)
	DeleteChangePasswordToken(tokenId int) error
	GetChangePasswordTokenByUserId(userId int) (*ChangePasswordToken, error)
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

func (adapter *ChangePasswordTokenPostgreAdapter) GetChangePasswordTokenByUserId(userId int) (*ChangePasswordToken, error) {
	return nil, nil
}

// Compile-time checks
var _ ChangePasswordTokenPort = (*ChangePasswordTokenPostgreAdapter)(nil)
