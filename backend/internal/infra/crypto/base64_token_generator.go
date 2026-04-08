package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"backend/internal/shared/crypto"

	"backend/internal/shared/config"
)

type MainTokenGenerator struct {
	hasher             crypto.SecretHasher
	decodedTokenLength int
	encoding           base64.Encoding
	tokenDuration      time.Duration
}

var _ crypto.SecurityTokenGenerator = (*MainTokenGenerator)(nil)

func NewMainTokenGenerator(
	hasher crypto.SecretHasher,
	cfg *config.Config,
) *MainTokenGenerator {
	return &MainTokenGenerator{
		hasher:             hasher,
		decodedTokenLength: int(cfg.TokenLength),
		encoding:           *base64.URLEncoding,
		tokenDuration:      time.Second * time.Duration(cfg.TokenDuration),
	}
}

func (generator *MainTokenGenerator) GenerateToken() (string, string, error) {
	byteNumber := generator.decodedTokenLength
	token := make([]byte, byteNumber) // Byte casuali
	// NOTA: rand.Read() manda programma in crash se non riesce a scrivere
	rand.Read(token) //nolint:errcheck

	encodedToken := make([]byte, generator.encoding.EncodedLen(byteNumber))
	generator.encoding.Encode(encodedToken, token)
	encodedTokenString := string(encodedToken) // Token codificato in Base64

	hashedTokenString, err := generator.hasher.HashSecret(encodedTokenString)
	if err != nil {
		return "", "", err
	}

	return encodedTokenString, hashedTokenString, nil
}

func (generator *MainTokenGenerator) ExpiryFromNow() time.Time {
	return time.Now().Add(generator.tokenDuration)
}
