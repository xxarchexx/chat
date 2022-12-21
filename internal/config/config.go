package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Host string
	Port string
	*NatsConfig
}

type NatsConfig struct {
	Host string
	Port string
}

func (config *AppConfig) LoadConfig() {
	godotenv.Load(".env")
	config.Host = os.Getenv("HOST")
	config.Port = os.Getenv("PORT")

	natsConfig := &NatsConfig{}
	natsConfig.Host = os.Getenv("NATS_HOST")
	natsConfig.Port = os.Getenv("NATS_PORT")
	config.NatsConfig = natsConfig
}
