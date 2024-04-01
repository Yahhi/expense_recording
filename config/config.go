// Package config helps to describe the data required to make app work. In fact these are Postgres connection settings
// and Telegram Bot token
package config

import (
	"os"
)

type Config struct {
	PGHost     string
	PGPort     string
	PGAdmin    string
	PGPass     string
	PGDbname   string
	TgBotToken string
}

func GetConfigFromEnv() Config {
	cfg := Config{
		PGHost:     os.Getenv("PGHOST"),
		PGPort:     os.Getenv("PGPORT"),
		PGDbname:   os.Getenv("PGDBNAME"),
		PGAdmin:    os.Getenv("PGADMIN"),
		PGPass:     os.Getenv("PGPASS"),
		TgBotToken: os.Getenv("TGTOKEN"),
	}
	return cfg
}
