package config

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Nome dell'applicativo
	Name string `json:"NAME"`

	// Porta su cui aprire il backend
	Port string `json:"PORT"`

	// URL per il Cloud DB
	CloudDBUrl string `json:"CLOUD_DB_URL"`

	// Quale mail adapter utilizzare, può essere o "terminal" o "mailtrap"
	MailAdapter string `json:"MAIL_ADAPTER"`

	// Crypto ===========================================================================

	// Fattore di costo per algoritmo bcrypt
	BcryptCost stringInt `json:"BCRYPT_COST"`

	// Lunghezza in byte di un token di sicurezza
	TokenLength stringInt `json:"TOKEN_LENGTH"`

	// Durata di un token di sicurezza in secondi
	TokenDuration stringInt `json:"TOKEN_DURATION"`
}

type stringInt int

func (st *stringInt) UnmarshalJSON(b []byte) error {
	var item any
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}
	switch v := item.(type) {
	case int:
		*st = stringInt(v)
	case float64:
		*st = stringInt(int(v))
	case string:
		// here convert the string into an integer
		i, err := strconv.Atoi(v)
		if err != nil {
			return err // caso in cui stringa non è intero
		}
		*st = stringInt(i)
	}
	return nil
}

func ReadConfigFromEnv() (*Config, error) {
	envDict, err := godotenv.Read(".env")
	if err != nil {
		return nil, fmt.Errorf("impossibile leggere file .env: %v", err)
	}
	jsonBody, err := json.Marshal(&envDict)
	if err != nil {
		return nil, fmt.Errorf("errore di marshaling contenuti .env: %v", err)
	}

	var config Config
	if err := json.Unmarshal(jsonBody, &config); err != nil {
		return nil, fmt.Errorf("errore di unmarshaling contenuti .env: %v", err)
	}
	return &config, nil
}
