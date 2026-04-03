package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

type Config struct {

	/*
		URL su cui si trova il front-end dell'applicativo. Viene usato dal package email
		per l'invio dei token di conferma/cambio password
	*/
	AppURL string `json:"APP_URL"`

	// Porta su cui aprire il backend
	Port string `json:"PORT"`

	// Quale mail adapter utilizzare, può essere o "terminal" o "smtp"
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

	// Cloud DB =========================================================================

	CloudDBHost     string    `json:"CLOUD_POSTGRES_HOST"`     // Host del Cloud DB
	CloudDBPort     stringInt `json:"CLOUD_POSTGRES_PORT"`     // Porta del Cloud DB
	CloudDBUser     string    `json:"CLOUD_POSTGRES_USER"`     // Nome utente per accedere a Cloud DB
	CloudDBPassword string    `json:"CLOUD_POSTGRES_PASSWORD"` // Password per accedere a Cloud DB
	CloudDBName     string    `json:"CLOUD_POSTGRES_DB"`     // Nome del Cloud DB

	// Sensor DB =========================================================================

	SensorDBHost     string    `json:"POSTGRES_HOST"`     // Host del Sensor DB
	SensorDBPort     stringInt `json:"POSTGRES_PORT"`     // Porta del Sensor DB
	SensorDBUser     string    `json:"POSTGRES_USER"`     // Nome utente per accedere a Sensor DB
	SensorDBPassword string    `json:"POSTGRES_PASSWORD"` // Password per accedere a Sensor DB
	SensorDBName     string    `json:"POSTGRES_DB"`     // Nome del Sensor DB

	// SMTP =========================================================================

	SMTPHost string    `json:"SMTP_HOST"` // Hostname dell'URL SMTP 
	SMTPPort stringInt `json:"SMTP_PORT"` // Numero porta URL SMTP 
	SMTPUser string    `json:"SMTP_USER"` // Nome utente SMTP 
	SMTPPass string    `json:"SMTP_PASS"` // Password SMTP 
	SMTPFrom string    `json:"SMTP_FROM"` // Indirizzo email da cui inviare email tramite SMTP 
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
	 envDict := map[string]string{}
	
	configType := reflect.TypeFor[Config]()
	
	// Itera su tutti i campi dello struct Config usando reflection
	for field := range configType.Fields() {
		envKey := field.Tag.Get("json")
		if envKey == "" { continue }

		value, ok := os.LookupEnv(envKey)
		if !ok {
			return nil, fmt.Errorf("cannot find field '%v' in env", envKey)
		}
		envDict[envKey] = value
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
