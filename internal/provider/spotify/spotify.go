package spotify

import (
	"context"

	sdk "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetMyPlaylists(ctx context.Context, token *oauth2.Token) ([]sdk.SimplePlaylist, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := sdk.New(httpClient)

	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	return playlists.Playlists, nil
}
