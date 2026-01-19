package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	BaseURL          string
	SessionSecret    string
	YoutubeTokenFile string
	SpotifyTokenFile string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "random-session-secret"
	}

	return &Config{
		Port:             port,
		BaseURL:          baseURL,
		SessionSecret:    sessionSecret,
		YoutubeTokenFile: "token_youtube.json",
		SpotifyTokenFile: "token_spotify.json",
	}
}
