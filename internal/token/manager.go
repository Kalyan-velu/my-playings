package token

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

// SaveToken saves an OAuth2 token to the specified file path
func SaveToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// LoadToken loads an OAuth2 token from the specified file path
func LoadToken(path string) (*oauth2.Token, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}
