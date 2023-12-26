package main

import (
	"log"
	"net"
	"sync"

	pb "github.com/grpc-chat/proto"
	"google.golang.org/grpc"
)

type GrpcChat struct {
	pb.GrpcChatServer
	activeClients map[string]pb.GrpcChat_ReceiveAllMsgsServer
	mu            sync.Mutex
}

// pb.UnimplementedGrpcChatServer

var tcpListenAddress = "0.0.0.0:50051"

func main() {
	log.Println("Server code is coming !!")

	listeningAddress, err := net.Listen("tcp", tcpListenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	// var _ pb.GrpcChatServer = &GrpcChat{}

	pb.RegisterGrpcChatServer(grpcServer, NewServer())

	log.Printf("Grpc chat server is running at %v\n", tcpListenAddress)

	if err := grpcServer.Serve(listeningAddress); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
