package crypto

import (
	"crypto/sha512"

	"backend/internal/shared/config"

	"golang.org/x/crypto/bcrypt"
)

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

type BcryptHasher struct {
	cost int
}

var _ SecretHasher = (*BcryptHasher)(nil) // Compile-time check

func NewBcryptHasher(cfg *config.Config) *BcryptHasher {
	return &BcryptHasher{
		cost: int(cfg.BcryptCost),
	}
}

/*
Hash con SHA-512 come workaround per limite di 72 byte di bcrypt.
NOTA: non rende il sistema necessariamente più sicuro, serve solo per avere un input
di lunghezza <= 72 byte
*/
func (h *BcryptHasher) preHash(plaintext string) []byte {
	plainPreHashed := sha512.Sum512([]byte(plaintext))
	return plainPreHashed[:]
}

func (h *BcryptHasher) HashSecret(plaintext string) (string, error) {
	preHashedPlain := h.preHash(plaintext)
	hash, err := bcrypt.GenerateFromPassword(preHashedPlain, h.cost)
	hashString := string(hash)
	if err != nil {
		return hashString, err
	}
	return hashString, nil
}

func (h *BcryptHasher) CompareHashAndSecret(hashed string, plaintext string) error {
	preHashedPlain := h.preHash(plaintext)
	return bcrypt.CompareHashAndPassword([]byte(hashed), preHashedPlain)
}
