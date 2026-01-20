package auth

import (
	"context"
	"my-playings/internal/config"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/spotify"
	"google.golang.org/api/oauth2/v2"
	yt "google.golang.org/api/youtube/v3"
)

type Auth struct {
	Store *sessions.CookieStore
}

func NewAuth(cfg *config.Config) *Auth {

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	googleProvider := google.New(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.BaseURL+"/auth/google/callback",
		yt.YoutubeReadonlyScope,
		oauth2.UserinfoEmailScope, oauth2.UserinfoProfileScope,
	)
	googleProvider.SetAccessType("offline")
	googleProvider.SetPrompt("consent select_account")
	spotifyProvider := spotify.New(
		cfg.SpotifyClientID,
		cfg.SpotifyClientSecret,
		cfg.BaseURL+"/auth/spotify/callback",
		"user-read-private", "playlist-read-private",
	)

	goth.UseProviders(googleProvider, spotifyProvider)

	return &Auth{Store: store}
}

func (a *Auth) BeginAuth(w http.ResponseWriter, r *http.Request, provider string) {
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	gothic.BeginAuthHandler(w, r)
}

func (a *Auth) CompleteAuth(w http.ResponseWriter, r *http.Request, provider string) (goth.User, error) {
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	return gothic.CompleteUserAuth(w, r)
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request, provider string) {
	if provider != "" {
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	}
	gothic.Logout(w, r)
}
