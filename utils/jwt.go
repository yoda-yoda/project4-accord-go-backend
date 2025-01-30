package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type CustomClaims struct {
	UserID    string `json:"sub"`
	Role      string `json:"role"`
	IsRefresh bool   `json:"isRefresh"`
	jwt.RegisteredClaims
}

type PublicKeyStore struct {
	keys map[string]*rsa.PublicKey
	mu   sync.RWMutex
}

func NewPublicKeyStore() *PublicKeyStore {
	return &PublicKeyStore{
		keys: make(map[string]*rsa.PublicKey),
	}
}

func (store *PublicKeyStore) LoadKeys(dir string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".pem" {
			continue
		}

		filename := file.Name()
		var kid string
		_, err := fmt.Sscanf(filename, "%s_public.pem", &kid)
		if err != nil {
			continue // 파일명 형식이 맞지 않으면 무시
		}

		path := filepath.Join(dir, filename)
		pubKeyData, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read public key file %s: %v", path, err)
		}

		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
		if err != nil {
			return fmt.Errorf("failed to parse public key from file %s: %v", path, err)
		}

		store.keys[kid] = pubKey
	}

	return nil
}

func (store *PublicKeyStore) GetPublicKey(kid string) (*rsa.PublicKey, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	pubKey, exists := store.keys[kid]
	if !exists {
		return nil, errors.New("public key not found for kid: " + kid)
	}
	return pubKey, nil
}

var (
	PublicKeyStoreInstance *PublicKeyStore
	once                   sync.Once
)

func InitializePublicKeyStore() {
	once.Do(func() {
		PublicKeyStoreInstance = NewPublicKeyStore()

		err := godotenv.Load()
		if err != nil {

		}

		publicKeyDir := os.Getenv("JWT_PUBLIC_KEY_DIR")
		if publicKeyDir == "" {
			publicKeyDir = "keys"
		}

		err = PublicKeyStoreInstance.LoadKeys(publicKeyDir)
		if err != nil {
			panic("Failed to load public keys: " + err.Error())
		}
	})
}

func ParseJWT(tokenString string) (*CustomClaims, error) {

	InitializePublicKeyStore()

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return nil, err
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("kid not found in token header")
	}

	pubKey, err := PublicKeyStoreInstance.GetPublicKey(kid)
	if err != nil {
		return nil, err
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(*CustomClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
