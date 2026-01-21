package app

import (
	"encoding/json"
	"fmt"
	"log"
	"my-playings/internal/auth"
	"my-playings/internal/config"
	"my-playings/internal/provider/spotify"
	"my-playings/internal/provider/youtube"
	tokenstore "my-playings/internal/token"
	"net/http"
	"sync"
	"time"
)

type ProviderName string

const (
	ProviderYoutube ProviderName = "google"
	ProviderSpotify ProviderName = "spotify"
)

type Server struct {
	cfg        *config.Config
	auth       *auth.Auth
	youtube    *youtube.Provider
	spotify    *spotify.Provider
	tokenStore *tokenstore.TokenStore
	mu         sync.RWMutex
}

func NewServer(cfg *config.Config, auth *auth.Auth, tokenStore *tokenstore.TokenStore, youtube *youtube.Provider, spotifyProvider *spotify.Provider) *Server {

	return &Server{
		cfg:        cfg,
		auth:       auth,
		tokenStore: tokenStore,
		youtube:    youtube,
		spotify:    spotifyProvider,
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("→", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Println("←", time.Since(start))
	})
}

func (s *Server) Routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/auth/{provider}", s.auth.HandleGothAuth)
	mux.HandleFunc("/auth/{provider}/callback", s.auth.HandleGothCallback)

	mux.HandleFunc("/youtube/playlists", s.handlePlaylists)
	mux.HandleFunc("/spotify/playlists", s.handleSpotifyPlaylists)
	return LoggingMiddleware(mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	var html = `<html><body>
		<p><a href="/auth/google">LogIn with Google (YouTube)</a></p>
		<p><a href="/auth/spotify">LogIn with Spotify</a></p>
		<hr>
		<p><a href="/youtube/playlists">View YouTube Playlists</a></p>
		<p><a href="/spotify/playlists">View Spotify Playlists</a></p>
	</body></html>`
	fmt.Fprint(w, html)
}

func (s *Server) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	items, err := s.youtube.GetMyPlayLists(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Error encoding playlists: %v", err)
	}
}

func (s *Server) handleSpotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	items, err := s.spotify.GetMyPlaylists(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Error encoding playlists: %v", err)
	}
}
