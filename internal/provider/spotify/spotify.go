package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type Service struct {
	Config *oauth2.Config
}

func NewService(baseURL string) (*Service, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET environment variables are not set")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     oauth2.Endpoint{AuthURL: "https://accounts.spotify.com/authorize", TokenURL: "https://accounts.spotify.com/api/token"},
		RedirectURL:  baseURL + "/auth/spotify/callback",
		Scopes:       []string{"user-read-private", "playlist-read-private"},
	}

	return &Service{
		Config: config,
	}, nil
}

func (s *Service) SaveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}

func (s *Service) LoadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (s *Service) GetMyPlaylists(ctx context.Context, token *oauth2.Token) ([]sdk.SimplePlaylist, error) {
	httpClient := s.Config.Client(ctx, token)
	client := sdk.New(httpClient)

	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	return playlists.Playlists, nil
}
