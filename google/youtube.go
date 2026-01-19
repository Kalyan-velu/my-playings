package youtube_playlists

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeService struct {
	Config *oauth2.Config
}

// NewYoutubeService creates a new YoutubeService from the client secret file
func NewYoutubeService(clientSecretDir string) (*YoutubeService, error) {
	files, err := os.ReadDir(clientSecretDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var clientSecretFile string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "client_secret") && strings.HasSuffix(file.Name(), ".json") {
			clientSecretFile = filepath.Join(clientSecretDir, file.Name())
			break
		}
	}

	if clientSecretFile == "" {
		return nil, fmt.Errorf("client_secret.json file not found")
	}

	data, err := os.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(data, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	return &YoutubeService{Config: config}, nil
}

func (s *YoutubeService) GetMyPlayLists(ctx context.Context, token *oauth2.Token) ([]*youtube.Playlist, error) {
	httpClient := s.Config.Client(ctx, token)
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

func (s *YoutubeService) SaveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func (s *YoutubeService) LoadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}
