package server

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "go-server/pkg/keyrotation"
	"go-server/utils"
)

func RunGRPCServer(store *utils.PublicKeyStore) error {
	lis, err := net.Listen("tcp", ":50051") // gRPC 서버 포트
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterKeyRotationNotifyServiceServer(s, NewKeyRotationNotifyServer(store))

	log.Println("Starting gRPC server on port 50051...")
	return s.Serve(lis)
}
