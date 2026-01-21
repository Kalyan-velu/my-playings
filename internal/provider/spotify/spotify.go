package spotify

import (
	"context"
	"fmt"
	"my-playings/internal/config"
	musicprovider "my-playings/internal/provider"
	token_store "my-playings/internal/token"

	sdk "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

type Provider struct {
	Config   *oauth2.Config
	provider *musicprovider.MusicProvider
}

func NewSpotifyProvider(cfg *config.Config, store *token_store.TokenStore) *Provider {
	spotifyConfig := &oauth2.Config{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURL:  cfg.BaseURL + "/auth/spotify/callback",
		Scopes:       []string{"user-read-playback-state", "user-read-currently-playing", "playlist-read-private", "playlist-read-collaborative", "user-top-read", "playlist-modify-public", "playlist-modify-private"},
		Endpoint:     spotify.Endpoint,
	}
	spotifyClient := &musicprovider.MusicProvider{
		Name:       "spotify",
		Config:     spotifyConfig,
		TokenStore: store,
	}
	return &Provider{
		Config:   spotifyConfig,
		provider: spotifyClient,
	}
}

func (s *Provider) getClient(ctx context.Context) (*sdk.Client, error) {
	client, err := s.provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	return sdk.New(client), nil
}

func (s *Provider) GetMyPlaylists(ctx context.Context) ([]sdk.SimplePlaylist, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("get my playlists: %w", err)
	}
	return playlists.Playlists, nil
}

func (s *Provider) GetPlaylistTracks(ctx context.Context, playlistID string) ([]*sdk.FullTrack, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	tracks, err := client.GetTracks(ctx, []sdk.ID{sdk.ID(playlistID)})
	if err != nil {
		return nil, fmt.Errorf("get playlist tracks: %w", err)
	}
	return tracks, nil
}
