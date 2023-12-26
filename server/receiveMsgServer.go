package main

import (
	"io"
	"log"

	pb "github.com/grpc-chat/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (gc *GrpcChat) ReceiveAllMsgs(chatStream pb.GrpcChat_ReceiveAllMsgsServer) error {

	log.Print("New client request")
	clientID := ""
	clientSet := false
	for {
		msgReq, err := chatStream.Recv()
		if err == io.EOF {
			if clientID != "" {
				gc.RemoveClient(clientID)
			}
			break
		}
		if err != nil {
			e, ok := status.FromError(err)
			if ok && e.Code() == codes.Canceled {
				if clientID != "" {
					gc.RemoveClient(clientID)
				}
			} else {
				log.Printf("Failed to read the msgReq data fromm Client : %v", err)
			}
			return nil
		}
		log.Printf("Received message %v, from %v, type %v, for %v\n\n", msgReq.GetMessage(), msgReq.GetSenderID(), msgReq.GetMessageType(), msgReq.GetRoomOrRecipientID())
		clientID = msgReq.SenderID
		if !clientSet {
			gc.AddClient(clientID, chatStream)
			clientSet = true
		}
		switch msgReq.GetMessageType() {
		case pb.MessageTypeEnum_DIRECT:
			gc.sendDirectMsg(msgReq, clientID)
		case pb.MessageTypeEnum_GROUP:
			gc.sendGroupMsg(msgReq, clientID)
		case pb.MessageTypeEnum_ANONYMOUS:
			gc.sendAnonymousMsg(msgReq, clientID)
		}

	}
	return nil
}

func (gc *GrpcChat) sendDirectMsg(message *pb.ReceiveMsgsRequest, clientID string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	clientStream, ok := gc.activeClients[*message.RoomOrRecipientID]
	if ok {
		err := clientStream.Send(&pb.ReceiveMsgsResponse{
			MessageType: message.GetMessageType(),
			Message:     message.GetMessage(),
			SenderID:    clientID,
		})
		if err != nil {
			log.Fatalf("Failed to send the Direct message : %v", err)
		}
	}
}

func (gc *GrpcChat) sendGroupMsg(message *pb.ReceiveMsgsRequest, reqClientID string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	for clientID, clientStream := range gc.activeClients {
		if clientID == reqClientID {
			continue
		}
		clientStream.Send(&pb.ReceiveMsgsResponse{
			MessageType:       message.GetMessageType(),
			Message:           message.GetMessage(),
			RoomOrRecipientID: message.RoomOrRecipientID,
			SenderID:          reqClientID,
		})
	}
}

func (gc *GrpcChat) sendAnonymousMsg(message *pb.ReceiveMsgsRequest, reqClientID string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	for clientID, clientStream := range gc.activeClients {
		if clientID == reqClientID {
			continue
		}
		clientStream.Send(&pb.ReceiveMsgsResponse{
			MessageType: message.GetMessageType(),
			Message:     message.GetMessage(),
		})
	}
}
