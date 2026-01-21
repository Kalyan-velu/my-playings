package youtube

import (
	"context"
	"fmt"
	"my-playings/internal/config"
	musicprovider "my-playings/internal/provider"
	token_store "my-playings/internal/token"

	"github.com/markbates/goth/providers/google"
	"golang.org/x/oauth2"
	gc "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

type Provider struct {
	Config   *oauth2.Config
	provider *musicprovider.MusicProvider
}

func NewYoutubeProvider(cfg *config.Config, store *token_store.TokenStore) *Provider {
	youtubeConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.BaseURL + "/auth/google/callback",
		Scopes: []string{yt.YoutubeReadonlyScope,
			gc.UserinfoEmailScope, gc.UserinfoProfileScope},
		Endpoint: google.Endpoint,
	}
	ytClient := &musicprovider.MusicProvider{
		Name:       "google",
		Config:     youtubeConfig,
		TokenStore: store,
	}
	return &Provider{
		Config:   youtubeConfig,
		provider: ytClient,
	}
}

func (s *Provider) GetMyPlayLists(ctx context.Context) ([]*yt.Playlist, error) {
	client, err := s.provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	service, err := yt.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create youtube service: %w", err)
	}

	var allPlaylists []*yt.Playlist
	pageToken := ""
	for {
		call := service.Playlists.List([]string{"snippet", "contentDetails"}).Mine(true).MaxResults(50).PageToken(pageToken)
		res, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list playlists: %w", err)
		}
		allPlaylists = append(allPlaylists, res.Items...)
		pageToken = res.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return allPlaylists, nil
}
