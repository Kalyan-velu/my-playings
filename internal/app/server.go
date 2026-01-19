package app

import (
	"encoding/json"
	"fmt"
	"log"
	"my-playings/internal/auth"
	"my-playings/internal/config"
	"my-playings/internal/provider/spotify"
	"my-playings/internal/provider/youtube"
	"net/http"
	"strings"
	"sync"

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
	var ytToken *oauth2.Token
	if yt != nil {
		var err error
		ytToken, err = yt.LoadToken(cfg.YoutubeTokenFile)
		if err != nil {
			log.Printf("No existing YouTube token found or failed to load: %v", err)
		}
	}

	var spToken *oauth2.Token
	if sp != nil {
		var err error
		spToken, err = sp.LoadToken(cfg.SpotifyTokenFile)
		if err != nil {
			log.Printf("No existing Spotify token found or failed to load: %v", err)
		}
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
	return mux
}

func (s *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body>
		<p><a href="/auth/google">LogIn with Google (YouTube)</a></p>
		<p><a href="/auth/spotify">LogIn with Spotify</a></p>
		//<hr>
		//<p><a href="/youtube/playlists">View YouTube Playlists</a></p>
		//<p><a href="/spotify/playlists">View Spotify Playlists</a></p>
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
		// If user fetching fails (e.g. 401 on identity info), we might still have a valid token
		// in the session or in the error if Goth provides it.
		// However, CompleteAuth handles the full exchange.
		// If it fails because of FetchUser, we are in a tough spot with Goth.
		log.Printf("Goth callback error: %v", err)

		// Try to manually handle the code exchange if Goth fails due to user info
		if provider == "google" && strings.Contains(err.Error(), "401") {
			code := r.URL.Query().Get("code")
			if code != "" && s.youtubeService != nil {
				token, exchangeErr := s.youtubeService.Config.Exchange(r.Context(), code)
				if exchangeErr == nil {
					log.Printf("Successfully exchanged code manually after Goth failure")
					s.handlePostAuthToken(w, r, "google", token)
					return
				}
				log.Printf("Manual exchange also failed: %v", exchangeErr)
			}
		}

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

		if s.youtubeService != nil {
			err := s.youtubeService.SaveToken(s.cfg.YoutubeTokenFile, t)
			if err != nil {
				log.Printf("Warning: failed to save YouTube token: %v", err)
			}
		}
	} else if provider == "spotify" {
		s.mu.Lock()
		s.spToken = t
		s.mu.Unlock()

		if s.spotifyService != nil {
			err := s.spotifyService.SaveToken(s.cfg.SpotifyTokenFile, t)
			if err != nil {
				log.Printf("Warning: failed to save Spotify token: %v", err)
			}
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
	token := s.ytToken
	s.mu.RUnlock()

	if token == nil {
		http.Redirect(w, r, "/auth/google", http.StatusTemporaryRedirect)
		return
	}

	if s.youtubeService == nil {
		http.Error(w, "YouTube service not available", http.StatusInternalServerError)
		return
	}

	items, err := s.youtubeService.GetMyPlayLists(r.Context(), token)
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
	token := s.spToken
	s.mu.RUnlock()

	if token == nil {
		http.Redirect(w, r, "/auth/spotify", http.StatusTemporaryRedirect)
		return
	}

	if s.spotifyService == nil {
		http.Error(w, "Spotify service not available", http.StatusInternalServerError)
		return
	}

	items, err := s.spotifyService.GetMyPlaylists(r.Context(), token)
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
