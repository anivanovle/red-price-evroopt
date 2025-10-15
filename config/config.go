package config

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbUrl    string
	BotToken string
	Logger   *slog.Logger
}

func MustConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	c := &Config{
		DbUrl:    os.Getenv("DB_URL"),
		BotToken: os.Getenv("TG_TOKEN"),
		Logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}

	return c
}
