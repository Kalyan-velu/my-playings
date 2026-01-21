package auth

import (
	"fmt"
	"my-playings/internal/config"
	tokenstore "my-playings/internal/token"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/spotify"
	"golang.org/x/oauth2"
	gc "google.golang.org/api/oauth2/v2"
	yt "google.golang.org/api/youtube/v3"
)

type Auth struct {
	Store  *sessions.CookieStore
	tokens *tokenstore.TokenStore
}

func NewAuth(cfg *config.Config, tokens *tokenstore.TokenStore) *Auth {
	var secureCookie bool
	if cfg.Environment != "development" {
		secureCookie = true
	}
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = secureCookie
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	googleProvider := google.New(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.BaseURL+"/auth/google/callback",
		yt.YoutubeReadonlyScope,
		gc.UserinfoEmailScope, gc.UserinfoProfileScope,
	)
	googleProvider.SetAccessType("offline")
	googleProvider.SetPrompt("consent select_account")
	spotifyProvider := spotify.New(
		cfg.SpotifyClientID,
		cfg.SpotifyClientSecret,
		cfg.BaseURL+"/auth/spotify/callback",
		"user-read-private", "user-read-playback-state", "user-read-currently-playing", "playlist-read-private", "playlist-read-collaborative", "user-top-read", "playlist-modify-public", "playlist-modify-private",
	)

	goth.UseProviders(googleProvider, spotifyProvider)

	return &Auth{
		Store:  store,
		tokens: tokens,
	}
}

// HandleGothAuth URL should be /auth/{provider}
func (a *Auth) HandleGothAuth(w http.ResponseWriter, r *http.Request) {
	if _, err := gothic.CompleteUserAuth(w, r); err != nil {
		gothic.BeginAuthHandler(w, r)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (a *Auth) HandleGothCallback(w http.ResponseWriter, r *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := &oauth2.Token{
		AccessToken:  gothUser.AccessToken,
		RefreshToken: gothUser.RefreshToken,
		Expiry:       gothUser.ExpiresAt,
	}

	if err := a.tokens.SaveToken(r.PathValue("provider"), token); err != nil {
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Successfully authenticated with %s\n", r.PathValue("provider"))
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
