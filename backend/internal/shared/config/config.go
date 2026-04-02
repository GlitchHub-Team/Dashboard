package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var configFields = map[string]string{
	"PORT":                "",
	"MAIL_ADAPTER":        "",
	"TOKEN_DURATION":      "",
	"TOKEN_LENGTH":        "",
	"BCRYPT_COST":         "",
	"AUTH_TOKEN_DURATION": "",
	"AUTH_TOKEN_SECRET":   "",
}

type Config struct {
	// Nome dell'applicativo
	Name string `json:"NAME"`

	/*
		URL su cui si trova il front-end dell'applicativo. Viene usato dal package email
		per l'invio dei token di conferma/cambio password
	*/
	AppURL string `json:"APP_URL"`

	// Porta su cui aprire il backend
	Port string `json:"PORT"`

	// Quale mail adapter utilizzare, può essere o "terminal" o "mailtrap"
	MailAdapter string `json:"MAIL_ADAPTER"`

	// Crypto ===========================================================================

	// Fattore di costo per algoritmo bcrypt
	BcryptCost stringInt `json:"BCRYPT_COST"`

	// Lunghezza in byte di un token di sicurezza
	TokenLength stringInt `json:"TOKEN_LENGTH"`

	// Durata di un token di sicurezza in secondi
	TokenDuration stringInt `json:"TOKEN_DURATION"`

	// Durata di un token di autenticazione in secondi
	AuthTokenDuration stringInt `json:"AUTH_TOKEN_DURATION"`

	/*
		Secret per fare firma di token di autenticazione.
		Dev'essere codificato in base 64 URL-SAFE SENZA PADDING (base64.RawURLEncoding) ed essere lungo 512 bit!

		NOTA: Ha lunghezza 512 bit da decoded, la codifica base 64 ha lunghezza maggiore
	*/
	AuthTokenSecret string `json:"AUTH_TOKEN_SECRET"`

	// SMTP =========================================================================

	/* Hostname dell'URL SMTP */
	SMTPHost string `json:"SMTP_HOST"`

	/* Numero porta URL SMTP */
	SMTPPort stringInt `json:"SMTP_PORT"`

	/* Nome utente SMTP */
	SMTPUser string `json:"SMTP_USER"`

	/* Password SMTP */
	SMTPPass string `json:"SMTP_PASS"`

	/* Indirizzo email da cui inviare email tramite SMTP */
	SMTPFrom string `json:"SMTP_FROM"`
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
		for key := range configFields {
			envDict[key] = os.Getenv(key)
		}
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
