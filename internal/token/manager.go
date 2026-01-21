package token_store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"golang.org/x/oauth2"
)

type TokenStore struct {
	Mu            sync.RWMutex
	EncryptionKey [32]byte // Store this in environment variable!
	FilePath      string
}

type StoredTokens struct {
	Providers map[string]*oauth2.Token `json:"providers"`
}

func NewTokenStore(filePath string, encryptionKey string) (*TokenStore, error) {
	var key [32]byte
	copy(key[:], encryptionKey)

	return &TokenStore{
		EncryptionKey: key,
		FilePath:      filePath,
	}, nil
}

func (ts *TokenStore) SaveToken(provider string, token *oauth2.Token) error {
	ts.Mu.Lock()
	defer ts.Mu.Unlock()

	tokens, _ := ts.loadTokensUnsafe()
	if tokens.Providers == nil {
		tokens.Providers = make(map[string]*oauth2.Token)
	}

	tokens.Providers[provider] = token

	data, err := json.Marshal(tokens)

	if err != nil {
		return err
	}

	encrypted, err := ts.encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(ts.FilePath, encrypted, 0600) // Restricted permissions
}

func (ts *TokenStore) GetToken(provider string) (*oauth2.Token, error) {
	ts.Mu.RLock()
	defer ts.Mu.RUnlock()

	tokens, err := ts.loadTokensUnsafe()
	if err != nil {
		return nil, err
	}

	token, ok := tokens.Providers[provider]
	if !ok {
		return nil, errors.New("token not found")
	}

	return token, nil
}

// TODO - Use SQLite or something else to store tokens
func (ts *TokenStore) loadTokensUnsafe() (*StoredTokens, error) {
	encrypted, err := os.ReadFile(ts.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &StoredTokens{}, nil
		}
		return nil, err
	}

	decrypted, err := ts.decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	var tokens StoredTokens
	err = json.Unmarshal(decrypted, &tokens)
	return &tokens, err
}

func (ts *TokenStore) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(ts.EncryptionKey[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (ts *TokenStore) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(ts.EncryptionKey[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
