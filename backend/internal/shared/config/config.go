package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
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
	BcryptCost StringInt `json:"BCRYPT_COST"`

	// Lunghezza in byte di un token di sicurezza
	TokenLength StringInt `json:"TOKEN_LENGTH"`

	// Durata di un token di sicurezza in secondi
	TokenDuration StringInt `json:"TOKEN_DURATION"`

	// Durata di un token di autenticazione in secondi
	AuthTokenDuration StringInt `json:"AUTH_TOKEN_DURATION"`

	/*
		Secret per fare firma di token di autenticazione.
		Dev'essere codificato in base 64 URL-SAFE SENZA PADDING (base64.RawURLEncoding) ed essere lungo 512 bit!

		NOTA: Ha lunghezza 512 bit da decoded, la codifica base 64 ha lunghezza maggiore
	*/
	AuthTokenSecret string `json:"AUTH_TOKEN_SECRET"`

	// Cloud DB =========================================================================

	CloudDBHost     string    `json:"CLOUD_POSTGRES_HOST"`     // Host del Cloud DB
	CloudDBPort     StringInt `json:"CLOUD_POSTGRES_PORT"`     // Porta del Cloud DB
	CloudDBUser     string    `json:"CLOUD_POSTGRES_USER"`     // Nome utente per accedere a Cloud DB
	CloudDBPassword string    `json:"CLOUD_POSTGRES_PASSWORD"` // Password per accedere a Cloud DB
	CloudDBName     string    `json:"CLOUD_POSTGRES_DB"`       // Nome del Cloud DB
	CloudDBTest     bool      // True se si usa il Cloud DB di test temporaneo. NOTA: Questa variabile non si può impostare tramite ENV.

	// Sensor DB =========================================================================

	SensorDBHost     string    `json:"POSTGRES_HOST"`     // Host del Sensor DB
	SensorDBPort     StringInt `json:"POSTGRES_PORT"`     // Porta del Sensor DB
	SensorDBUser     string    `json:"POSTGRES_USER"`     // Nome utente per accedere a Sensor DB
	SensorDBPassword string    `json:"POSTGRES_PASSWORD"` // Password per accedere a Sensor DB
	SensorDBName     string    `json:"POSTGRES_DB"`       // Nome del Sensor DB
	SensorDBTest     bool      // True se si usa il Sensor DB di test temporaneo. NOTA: Questa variabile non si può impostare tramite ENV.

	// SMTP =========================================================================

	SMTPHost string    `json:"SMTP_HOST"` // Hostname dell'URL SMTP
	SMTPPort StringInt `json:"SMTP_PORT"` // Numero porta URL SMTP
	SMTPUser string    `json:"SMTP_USER"` // Nome utente SMTP
	SMTPPass string    `json:"SMTP_PASS"` // Password SMTP
	SMTPFrom string    `json:"SMTP_FROM"` // Indirizzo email da cui inviare email tramite SMTP
}

type StringInt int

func (st *StringInt) UnmarshalJSON(b []byte) error {
	var item any
	if err := json.Unmarshal(b, &item); err != nil {
		return err
	}
	switch v := item.(type) {
	case int:
		*st = StringInt(v)
	case float64:
		*st = StringInt(int(v))
	case string:
		// here convert the string into an integer
		i, err := strconv.Atoi(v)
		if err != nil {
			return err // caso in cui stringa non è intero
		}
		*st = StringInt(i)
	}
	return nil
}

func ReadConfigFromEnv(log *zap.Logger) (*Config, error) {
	// NOTA: Il file .env ha la priorità sulle variabili d'ambiente
	envDict, err := godotenv.Read(".env")
	if envDict == nil {
		envDict = make(map[string]string, 0)
	}
	if err != nil {
		log.Sugar().Infof("Cannot read env: %v", err)
	}
	if len(envDict) == 0 {
		log.Sugar().Infof(".env file is empty")
	}

	// Itera su tutti i campi dello struct Config usando reflection
	var missingFields []string
	for field := range reflect.TypeFor[Config]().Fields() {
		envKey := field.Tag.Get("json")
		if envKey == "" {
			continue
		}

		_, inEnvFile := envDict[envKey]
		value, inEnv := os.LookupEnv(envKey)

		if !inEnv && !inEnvFile {
			missingFields = append(missingFields, envKey)
			continue
		}

		// Inserisci in envDict se NON trovo già
		if !inEnvFile {
			envDict[envKey] = value
		}
	}

	if missingFields != nil {
		return nil, fmt.Errorf("the following env variables are missing: %v", strings.Join(missingFields, ", "))
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
