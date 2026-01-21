package spotify

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"my-playings/internal/config"
	token_store "my-playings/internal/token"
)

func TestNewSpotifyProvider_UsesTokenStore(t *testing.T) {
	cfg := &config.Config{
		SpotifyClientID:     "client-id",
		SpotifyClientSecret: "client-secret",
		BaseURL:             "http://localhost:8080",
	}
	store, err := token_store.NewTokenStore(filepath.Join(t.TempDir(), "token.enc"), strings.Repeat("k", 32))
	if err != nil {
		t.Fatalf("create token store: %v", err)
	}

	provider := NewSpotifyProvider(cfg, store)
	if provider.provider == nil {
		t.Fatal("expected music provider to be initialized")
	}
	if provider.provider.TokenStore != store {
		t.Fatal("expected token store to be wired into music provider")
	}
	if provider.provider.Name != "spotify" {
		t.Fatalf("expected provider name spotify, got %s", provider.provider.Name)
	}
}

func TestGetMyPlaylistsWithoutTokenReturnsError(t *testing.T) {
	cfg := &config.Config{
		SpotifyClientID:     "client-id",
		SpotifyClientSecret: "client-secret",
		BaseURL:             "http://localhost:8080",
	}
	store, err := token_store.NewTokenStore(filepath.Join(t.TempDir(), "token.enc"), strings.Repeat("k", 32))
	if err != nil {
		t.Fatalf("create token store: %v", err)
	}

	provider := NewSpotifyProvider(cfg, store)
	if _, err := provider.GetMyPlaylists(context.Background()); err == nil {
		t.Fatal("expected error when token is missing")
	}
}
