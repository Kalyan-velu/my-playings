package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	playlists "my-playings/google"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

var (
	googleOauthConfig *oauth2.Config
	// TODO: Use a secure way to store tokens in production
	token *oauth2.Token
)

func init() {
	// Dynamically find the client secret file
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var clientSecretFile string
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 13 && file.Name()[:13] == "client_secret" {
			clientSecretFile = file.Name()
			break
		}
	}

	if clientSecretFile == "" {
		log.Fatal("lient_secret.jsonc file not found")
	}

	data, err := os.ReadFile(clientSecretFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(data, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	// Google's default redirect URI for web apps might need to be configured in console
	// For local testing, ensure it matches what's in the console (usually http://localhost:8080/callback)
	googleOauthConfig = config
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleMain)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/callback", handleCallback)
	mux.HandleFunc("/playlists", handlePlaylists)

	fmt.Println("Server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body><a href="/login">Google LogIn</a></body></html>`
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	t, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	token = t
	http.Redirect(w, r, "/playlists", http.StatusTemporaryRedirect)
}

func handlePlaylists(w http.ResponseWriter, r *http.Request) {
	if token == nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	items, err := playlists.GetMyPlayLists(context.Background(), client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
