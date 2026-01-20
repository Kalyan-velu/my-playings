package app

import (
	"encoding/json"
	"fmt"
	"log"
	"my-playings/internal/auth"
	"my-playings/internal/config"
	"my-playings/internal/provider/spotify"
	"my-playings/internal/provider/youtube"
	"my-playings/internal/token"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

type Server struct {
	cfg            *config.Config
	auth           *auth.Auth
	youtubeService *youtube.Service
	spotifyService *spotify.Service
	ytToken        *oauth2.Token
	spToken        *oauth2.Token
	mu             sync.RWMutex
}

func NewServer(cfg *config.Config, auth *auth.Auth, yt *youtube.Service, sp *spotify.Service) *Server {
	ytToken, err := token.LoadToken(cfg.YoutubeTokenFile)
	if err != nil {
		log.Printf("No existing YouTube token found or failed to load: %v", err)
	}

	spToken, err := token.LoadToken(cfg.SpotifyTokenFile)
	if err != nil {
		log.Printf("No existing Spotify token found or failed to load: %v", err)
	}

	return &Server{
		cfg:            cfg,
		auth:           auth,
		youtubeService: yt,
		spotifyService: sp,
		ytToken:        ytToken,
		spToken:        spToken,
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// before
		log.Println("→", r.Method, r.URL.Path)

		next.ServeHTTP(w, r) // pass control forward

		// after
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

	mux.HandleFunc("/youtube/playlists", s.handleYoutubePlaylists)
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

func (s *Server) handleYoutubePlaylists(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	ytToken := s.ytToken
	s.mu.RUnlock()

	fmt.Println(ytToken)
	if ytToken == nil {
		http.Redirect(w, r, "/auth/google", http.StatusTemporaryRedirect)
		return
	}

	if s.youtubeService == nil {
		http.Error(w, "YouTube service not available", http.StatusInternalServerError)
		return
	}

	items, err := s.youtubeService.GetMyPlayLists(r.Context(), ytToken)
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
	s.mu.RLock()
	spToken := s.spToken
	s.mu.RUnlock()

	if spToken == nil {
		http.Redirect(w, r, "/auth/spotify", http.StatusTemporaryRedirect)
		return
	}

	if s.spotifyService == nil {
		http.Error(w, "Spotify service not available", http.StatusInternalServerError)
		return
	}

	items, err := s.spotifyService.GetMyPlaylists(r.Context(), spToken)
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
