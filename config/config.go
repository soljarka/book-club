package config

import (
	"log"

	"github.com/joho/godotenv"
)

// Config describes configuration.
type Config struct {
	BotToken string `required:"true"`
}

func init() {
	log.Println("Loading .env")
	err := godotenv.Load()
	if err != nil {
		log.Print("Couldn't find .env file")
	}
}
