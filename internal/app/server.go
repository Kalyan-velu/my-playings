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

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
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

func NewServer(cfg *config.Config, tokenStore *tokenstore.TokenStore, youtube *youtube.Provider, spotifyProvider *spotify.Provider) *Server {
	authNew := auth.NewAuth(cfg)

	return &Server{
		cfg:        cfg,
		auth:       authNew,
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
	mux.HandleFunc("/", s.handleMain)
	mux.HandleFunc("/auth/", s.handleGothLogin)
	mux.HandleFunc("/auth/google/callback", s.handleGothCallback)
	mux.HandleFunc("/auth/spotify/callback", s.handleGothCallback)
	mux.HandleFunc("/logout/", s.handleLogout)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/youtube/playlists", s.handlePlaylists)
	mux.HandleFunc("/spotify/playlists", s.handleSpotifyPlaylists)
	return LoggingMiddleware(mux)
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body>
		<p><a href="/auth/google">LogIn with Google (YouTube)</a></p>
		<p><a href="/auth/spotify">LogIn with Spotify</a></p>
		<hr>
		<p><a href="/youtube/playlists">View YouTube Playlists</a></p>
		<p><a href="/spotify/playlists">View Spotify Playlists</a></p>
	</body></html>`
	fmt.Fprint(w, html)
}

func (s *Server) handleGothLogin(w http.ResponseWriter, r *http.Request) {
	provider := s.getProvider(r)
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	if user, err := s.auth.CompleteAuth(w, r, provider); err == nil {
		s.handlePostAuth(w, r, user)
	} else {
		s.auth.BeginAuth(w, r, provider)
	}
}

func (s *Server) handleGothCallback(w http.ResponseWriter, r *http.Request) {
	provider := s.getProvider(r)
	if provider == "" {
		// Fallback for callback paths
		provider = "google"
		if strings.Contains(r.URL.Path, "spotify") {
			provider = "spotify"
		}
	}

	user, err := s.auth.CompleteAuth(w, r, provider)
	if err != nil {
		log.Printf("Goth callback error: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return

	}
	s.handlePostAuth(w, r, user)
}

func (s *Server) handlePostAuth(w http.ResponseWriter, r *http.Request, user goth.User) {
	t := &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.ExpiresAt,
		TokenType:    "Bearer",
	}

	s.handlePostAuthToken(w, r, user.Provider, t)
}

func (s *Server) handlePostAuthToken(w http.ResponseWriter, r *http.Request, provider string, t *oauth2.Token) {
	if provider == "google" {
		s.mu.Lock()
		s.ytToken = t
		s.mu.Unlock()

		err := token.SaveToken(s.cfg.YoutubeTokenFile, t)
		if err != nil {
			log.Printf("Warning: failed to save YouTube token: %v", err)
		}
	} else if provider == "spotify" {
		s.mu.Lock()
		s.spToken = t
		s.mu.Unlock()

		err := token.SaveToken(s.cfg.SpotifyTokenFile, t)
		if err != nil {
			log.Printf("Warning: failed to save Spotify token: %v", err)
		}
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	provider := s.getProvider(r)
	s.auth.Logout(w, r, provider)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

func (s *Server) getProvider(r *http.Request) string {
	provider := strings.TrimPrefix(r.URL.Path, "/auth/")
	if strings.HasPrefix(r.URL.Path, "/logout/") {
		provider = strings.TrimPrefix(r.URL.Path, "/logout/")
	}
	if strings.Contains(provider, "/") {
		provider = strings.Split(provider, "/")[0]
	}
	if provider == "" {
		provider = r.URL.Query().Get("provider")
	}
	return provider
}
