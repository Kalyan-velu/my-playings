package auth

import (
	"my-playings/internal/config"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/spotify"
	gc "google.golang.org/api/oauth2/v2"
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

	return &Auth{Store: store}
}

// HandleGothAuth URL should be /auth/{provider}
func (a *Auth) HandleGothAuth(w http.ResponseWriter, r *http.Request) {
	if _, err := gothic.CompleteUserAuth(w, r); err != nil {
		gothic.BeginAuthHandler(w, r)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
