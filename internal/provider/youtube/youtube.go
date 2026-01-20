package youtube

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetMyPlayLists(ctx context.Context, token *oauth2.Token) ([]*youtube.Playlist, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	service, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create youtube service: %w", err)
	}

	var allPlaylists []*youtube.Playlist
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
