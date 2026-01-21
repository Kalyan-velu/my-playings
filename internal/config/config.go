package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port          string `env:"PORT,required" envDefault:"8080"`
	BaseURL       string `env:"BASE_URL,required" envDefault:"http://localhost:8080"`
	SessionSecret string `env:"SESSION_SECRET,required"`
	//Spotify
	SpotifyClientID     string `env:"SPOTIFY_CLIENT_ID,required"`
	SpotifyClientSecret string `env:"SPOTIFY_CLIENT_SECRET,required"`
	SpotifyTokenFile    string `env:"SPOTIFY_TOKEN_FILE" envDefault:"token_spotify.json"`
	//Google
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	YoutubeTokenFile   string `env:"YOUTUBE_TOKEN_FILE" envDefault:"token_youtube.json"`
	EncryptionKey      string `env:"ENCRYPTION_KEY,required"`
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg, err := env.ParseAs[Config]()

	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	return &cfg
}
