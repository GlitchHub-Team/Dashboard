package crypto


type SecretHasher interface {
	/* Ottieni hash di plaintext sottoforma di string*/
	HashSecret(plaintext string) (string, error)

	/*
		Controlla plaintext e hashed. E' IMPORTANTE USARE FUNZIONI SICURE DA TIMING ATTACKS
		come bcrypt.CompareHashAndPassword.

		Se il controllo passa, allora ritorna errore nil, altrimenti ritorna errore non-nil.
	*/
	CompareHashAndSecret(hashed string, plaintext string) error
}
