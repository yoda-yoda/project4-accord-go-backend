package server

import (
	"context"
	"fmt"
	"log"

	pb "go-server/pkg/keyrotation"
	"go-server/utils"
)

type KeyRotationNotifyServer struct {
	pb.UnimplementedKeyRotationNotifyServiceServer
	store *utils.PublicKeyStore
}

func NewKeyRotationNotifyServer(store *utils.PublicKeyStore) *KeyRotationNotifyServer {
	return &KeyRotationNotifyServer{
		store: store,
	}
}

func (s *KeyRotationNotifyServer) NotifyKeyRolled(ctx context.Context, req *pb.NotifyKeyRolledRequest) (*pb.NotifyKeyRolledResponse, error) {
	log.Printf("Received key rotation notification: prevKid=%s, currKid=%s, rolledAt=%s",
		req.GetPreviousKid(), req.GetCurrentKid(), req.GetRolledAt())

	if req.GetCurrentPublicKeyPem() == "" {
		return nil, fmt.Errorf("no public key pem provided")
	}

	err := s.store.AddOrUpdateKey(ctx, req.GetCurrentKid(), req.GetCurrentPublicKeyPem())
	if err != nil {
		return nil, fmt.Errorf("failed to add/update key in store: %v", err)
	}

	return &pb.NotifyKeyRolledResponse{
		Message: "Public key updated successfully.",
	}, nil
}
