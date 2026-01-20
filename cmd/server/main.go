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

	ytService := youtube.NewService()
	spService := spotify.NewService()
	authenticator := auth.NewAuth(cfg)

	server := app.NewServer(cfg, authenticator, ytService, spService)

	fmt.Printf("Server started at %s\n", cfg.Port)
	err := http.ListenAndServe(":"+cfg.Port, server.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
