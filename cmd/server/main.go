package main

import (
	"fmt"
	"log"
	"my-playings/internal/app"
	"my-playings/internal/config"
	"my-playings/internal/provider/spotify"
	"my-playings/internal/provider/youtube"
	tokenstore "my-playings/internal/token"
	"net/http"
)

func main() {
	cfg := config.Load()

	tokens, err := tokenstore.NewTokenStore("token.enc", cfg.EncryptionKey)
	if err != nil {
		log.Fatal(err)
	}

	yt := youtube.NewYoutubeProvider(cfg, tokens)
	spotifyProvider := spotify.NewSpotifyProvider(cfg, tokens)

	server := app.NewServer(cfg, tokens, yt, spotifyProvider)

	fmt.Printf("Server started at %s\n", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, server.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
