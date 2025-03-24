package config

import (
	"os"
	"strings"
)

// Config representa a configuração da aplicação
type Config struct {
	AwsRegion string
	AwsAccessKeyID string
	AwsSecretAccessKey string
	ServerPort string
}

// LoadConfig carrega as configurações do ambiente
func LoadConfig() *Config {
	return &Config{
		AwsRegion:          getEnv("AWS_REGION", "us-east-1"),
		AwsAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AwsSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv obtém uma variável de ambiente ou retorna o valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return value
}
