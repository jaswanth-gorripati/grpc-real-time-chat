package main

import (
	"log"

	pb "github.com/grpc-chat/proto"
)

func NewServer() *GrpcChat {
	return &GrpcChat{
		activeClients: make(map[string]pb.GrpcChat_ReceiveAllMsgsServer),
	}
}

func (gc *GrpcChat) AddClient(clientID string, clientStream pb.GrpcChat_ReceiveAllMsgsServer) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	log.Printf("Adding new Client %v\n", clientID)
	gc.activeClients[clientID] = clientStream
}

func (gc *GrpcChat) RemoveClient(clientID string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	log.Printf("Removing  Client %v\n", clientID)
	delete(gc.activeClients, clientID)
}
