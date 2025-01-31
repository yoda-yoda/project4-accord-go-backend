package utils

import (
	"crypto/rsa"
	"fmt"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

// PublicKeyStore는 KID와 RSA 공개키를 매핑하는 저장소입니다.
type PublicKeyStore struct {
	keys map[string]*rsa.PublicKey
	mu   sync.RWMutex
}

// NewPublicKeyStore는 새로운 PublicKeyStore 인스턴스를 생성합니다.
func NewPublicKeyStore() *PublicKeyStore {
	return &PublicKeyStore{
		keys: make(map[string]*rsa.PublicKey),
	}
}

// AddOrUpdateKey는 PEM 형식의 공개키를 파싱하여 저장소에 추가하거나 업데이트합니다.
func (store *PublicKeyStore) AddOrUpdateKey(kid, pemStr string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// PEM 파싱
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pemStr))
	if err != nil {
		return fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	// Map에 저장(기존 kid가 있어도 덮어씀)
	store.keys[kid] = pubKey
	return nil
}

func (store *PublicKeyStore) RemoveKey(kid string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.keys, kid)
}

func (store *PublicKeyStore) GetKey(kid string) (*rsa.PublicKey, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	key, exists := store.keys[kid]
	if !exists {
		return nil, fmt.Errorf("public key not found for kid: %s", kid)
	}
	return key, nil
}
