package server

import (
	"context"
	"fmt"
	"log"

	pb "go-server/pkg/keyrotation"
	"go-server/utils"
)

// KeyRotationNotifyServiceServer 구현체
type KeyRotationNotifyServer struct {
	pb.UnimplementedKeyRotationNotifyServiceServer
	// 필요하다면 Store에 대한 의존성 주입
	store *utils.PublicKeyStore
}

// NewKeyRotationNotifyServer : 생성자(혹은 간단한 빌더)
func NewKeyRotationNotifyServer(store *utils.PublicKeyStore) *KeyRotationNotifyServer {
	return &KeyRotationNotifyServer{
		store: store,
	}
}

// NotifyKeyRolled : 인증서버에서 키가 롤링된 후 호출될 RPC 메서드
func (s *KeyRotationNotifyServer) NotifyKeyRolled(ctx context.Context, req *pb.NotifyKeyRolledRequest) (*pb.NotifyKeyRolledResponse, error) {
	log.Printf("Received key rotation notification: prevKid=%s, currKid=%s, rolledAt=%s",
		req.GetPreviousKid(), req.GetCurrentKid(), req.GetRolledAt())

	// (1) 새로운 Public Key(PublicKeyPem) 검증 및 파싱
	if req.GetCurrentPublicKeyPem() == "" {
		return nil, fmt.Errorf("no public key pem provided")
	}

	err := s.store.AddOrUpdateKey(req.GetCurrentKid(), req.GetCurrentPublicKeyPem())
	if err != nil {
		return nil, fmt.Errorf("failed to add/update key in store: %v", err)
	}

	// (2) 필요 시, previousKid 제거 로직(또는 유지 로직)
	//     인증 서버 정책에 맞게 이전 키도 유지할지, 제거할지 결정
	//     이번 예시에서는 유지한다고 가정
	//     if req.GetPreviousKid() != "" {
	//         s.store.RemoveKey(req.GetPreviousKid())
	//     }

	return &pb.NotifyKeyRolledResponse{
		Message: "Public key updated successfully.",
	}, nil
}
