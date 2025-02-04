package utils

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

type PublicKeyStore struct {
	redisClient *redis.Client
	previousKid string
}

func NewPublicKeyStore(redisClient *redis.Client) *PublicKeyStore {
	return &PublicKeyStore{
		redisClient: redisClient,
	}
}

func MarshalRSAPublicKey(pubKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
	var pemBuffer bytes.Buffer
	pem.Encode(&pemBuffer, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	return pemBuffer.Bytes(), nil
}

func (store *PublicKeyStore) AddOrUpdateKey(ctx context.Context, kid, pemStr string) error {
	// PEM 파싱
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pemStr))
	if err != nil {
		return fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	// Redis에 저장
	pubKeyBytes, err := MarshalRSAPublicKey(pubKey)
	if err != nil {
		return fmt.Errorf("failed to marshal RSA public key: %w", err)
	}

	if store.previousKid != "" && store.previousKid != kid {
		err := store.RemoveKey(ctx, store.previousKid)
		if err != nil {
			return fmt.Errorf("failed to remove previous key from store: %w", err)
		}
	}

	err = store.redisClient.Set(ctx, kid, pubKeyBytes, 0).Err()
	if err != nil {
		return err
	}

	store.previousKid = kid

	return nil
}

func (store *PublicKeyStore) RemoveKey(ctx context.Context, kid string) error {
	return store.redisClient.Del(ctx, kid).Err()
}

func (store *PublicKeyStore) GetKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := store.redisClient.Get(ctx, kid).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}
		return nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return pubKey, nil
}
