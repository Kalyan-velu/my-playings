package playlists

import (
	"context"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func GetMyPlayLists(ctx context.Context, httpClient *http.Client) ([]*youtube.Playlist, error) {
	service, err := youtube.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	var allPlaylists []*youtube.Playlist
	pageToken := ""
	for {
		call := service.Playlists.List([]string{"snippet", "contentDetails"}).Mine(true).MaxResults(50).PageToken(pageToken)
		res, err := call.Do()
		if err != nil {
			return nil, err
		}
		allPlaylists = append(allPlaylists, res.Items...)
		pageToken = res.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return allPlaylists, nil
}
