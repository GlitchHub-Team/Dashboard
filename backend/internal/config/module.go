package config

import (
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

type Config struct {
	Name        string `json:"NAME"`
	Port        string `json:"PORT"`
	CloudDBUrl  string `json:"CLOUD_DB_URL"`
	MailAdapter string `json:"MAIL_ADAPTER"`
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

var Module = fx.Module(
	"config",
	fx.Provide(ReadConfigFromEnv),
)
