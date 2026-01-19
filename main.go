package main

import (
	"encoding/json"
	"fmt"
	"log"
	playlists "my-playings/google"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"
)

type Server struct {
	youtubeService *playlists.YoutubeService
	token          *oauth2.Token
	tokenFile      string
	mu             sync.RWMutex // For thread-safe token access
}

func main() {
	ytService, err := playlists.NewYoutubeService(".")
	if err != nil {
		log.Fatalf("Failed to initialize YouTube service: %v", err)
	}

	tokenFile := "token.json"
	initialToken, err := ytService.LoadToken(tokenFile)
	if err != nil {
		log.Printf("No existing token found or failed to load: %v", err)
	}

	s := &Server{
		youtubeService: ytService,
		token:          initialToken,
		tokenFile:      tokenFile,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleMain)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/callback", s.handleCallback)
	mux.HandleFunc("/playlists", s.handlePlaylists)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started at http://localhost:%s\n", port)
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body><a href="/login">Google LogIn</a></body></html>`
	fmt.Fprint(w, html)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// In production, use a secure, random state token and verify it in the callback
	state := "random-state-token"
	url := s.youtubeService.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	t, err := s.youtubeService.Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.token = t
	s.mu.Unlock()

	err = s.youtubeService.SaveToken(s.tokenFile, t)
	if err != nil {
		log.Printf("Warning: failed to save token: %v", err)
	}

	http.Redirect(w, r, "/playlists", http.StatusTemporaryRedirect)
}

func (s *Server) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	token := s.token
	s.mu.RUnlock()

	if token == nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	items, err := s.youtubeService.GetMyPlayLists(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// The token might have been refreshed. oauth2.Config.Client returns a client
	// that handles refreshing. We should check if the token changed and save it.
	// However, the standard oauth2 library's TokenSource doesn't easily expose if it refreshed
	// unless we use a custom TokenSource with a notify function.
	// For now, simple implementation is enough as requested.

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Error encoding playlists: %v", err)
	}
}
