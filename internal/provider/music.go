package musicprovider

import (
	"context"
	"fmt"
	store "my-playings/internal/token"
	"net/http"

	"golang.org/x/oauth2"
)

type MusicProvider struct {
	Name       string
	Config     *oauth2.Config
	TokenStore *store.TokenStore
}

func (p *MusicProvider) GetClient(ctx context.Context) (*http.Client, error) {
	token, err := p.TokenStore.GetToken(p.Name)
	if err != nil {
		return nil, err
	}
	// Create token source that auto-refreshes
	tokenSource := p.Config.TokenSource(ctx, token)

	// Wrap to save refreshed tokens
	wrappedSource := &savingTokenSource{
		base:       tokenSource,
		provider:   p,
		tokenStore: p.TokenStore,
	}

	return oauth2.NewClient(ctx, wrappedSource), nil
}

// savingTokenSource wraps oauth2.TokenSource to save refreshed tokens
type savingTokenSource struct {
	base       oauth2.TokenSource
	provider   *MusicProvider
	tokenStore *store.TokenStore
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		return nil, err
	}

	// Do this asynchronously to avoid blocking
	go func() {
		if err := s.tokenStore.SaveToken(s.provider.Name, token); err != nil {
			fmt.Printf("Error saving refreshed token for %s: %v\n", s.provider.Name, err)
		}
	}()

	return token, nil
}
