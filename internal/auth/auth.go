package auth

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/spotify"
	oauth2 "google.golang.org/api/oauth2/v2"
	yt "google.golang.org/api/youtube/v3"
)

type Auth struct {
	Store *sessions.CookieStore
}

func NewAuth(sessionSecret, baseURL string, googleClientID, googleClientSecret, spotifyClientID, spotifyClientSecret string) *Auth {
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false // Set to true in production
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	var providers []goth.Provider

	if googleClientID != "" && googleClientSecret != "" {
		providers = append(providers, google.New(
			googleClientID,
			googleClientSecret,
			baseURL+"/auth/google/callback",
			yt.YoutubeReadonlyScope,
			oauth2.UserinfoEmailScope, oauth2.UserinfoProfileScope,
		))
	}

	if spotifyClientID != "" && spotifyClientSecret != "" {
		providers = append(providers, spotify.New(
			spotifyClientID,
			spotifyClientSecret,
			baseURL+"/auth/spotify/callback",
			"user-read-private", "playlist-read-private",
		))
	}

	goth.UseProviders(providers...)

	return &Auth{Store: store}
}

func (a *Auth) GetProviderFromRequest(r *http.Request) string {
	// Simple provider extraction from URL
	return ""
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
