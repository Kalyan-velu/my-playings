package main

import (
	"fmt"
	"log"
	"my-playings/internal/app"
	"my-playings/internal/auth"
	"my-playings/internal/config"
	"my-playings/internal/provider/spotify"
	"my-playings/internal/provider/youtube"
	"net/http"
)

func main() {
	cfg := config.Load()

	ytService, err := youtube.NewService(".")
	if err != nil {
		log.Printf("Warning: YouTube service initialization failed: %v", err)
	}

	spService, err := spotify.NewService(cfg.BaseURL)
	if err != nil {
		log.Printf("Warning: Spotify service initialization failed: %v", err)
	}

	var googleClientID, googleClientSecret string
	if ytService != nil && ytService.Config != nil {
		googleClientID = ytService.Config.ClientID
		googleClientSecret = ytService.Config.ClientSecret
	}

	var spotifyClientID, spotifyClientSecret string
	if spService != nil && spService.Config != nil {
		spotifyClientID = spService.Config.ClientID
		spotifyClientSecret = spService.Config.ClientSecret
	}

	authenticator := auth.NewAuth(
		cfg.SessionSecret,
		cfg.BaseURL,
		googleClientID,
		googleClientSecret,
		spotifyClientID,
		spotifyClientSecret,
	)

	server := app.NewServer(cfg, authenticator, ytService, spService)

	fmt.Printf("Server started at %s\n", cfg.BaseURL)
	err = http.ListenAndServe(":"+cfg.Port, server.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
