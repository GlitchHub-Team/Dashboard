package crypto

//go:generate mockgen -destination=../../../tests/shared/crypto/mocks/secret_hasher.go -package=mocks . SecretHasher

type SecretHasher interface {
	/* Ottieni hash di plaintext sottoforma di string*/
	HashSecret(plaintext string) (string, error)

	/*
		Controlla che plaintext e hashed siano uguali. E' IMPORTANTE USARE FUNZIONI SICURE DA TIMING ATTACKS
		come bcrypt.CompareHashAndPassword.

		Se il controllo passa, allora ritorna errore nil, altrimenti ritorna errore non-nil.
	*/
	CompareHashAndSecret(hashed string, plaintext string) error
}
