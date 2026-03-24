package auth

type passwordTokenPostgreRepository struct{}

func newPasswordTokenPostgreRepository() *passwordTokenPostgreRepository {
	return &passwordTokenPostgreRepository{}
}

type confirmTokenPostgreRepository struct{}

func newConfirmTokenPostgreRepository() *confirmTokenPostgreRepository {
	return &confirmTokenPostgreRepository{}
}
