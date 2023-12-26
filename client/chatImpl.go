package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	pb "github.com/grpc-chat/proto"
	"github.com/rivo/tview"
)

const (
	stateChooseOption = iota
	stateRoomName
	stateEnterUsername
	stateEnterMessage
)

var inputState int = stateChooseOption

func connectToChatServer(chatClient pb.GrpcChatClient, userID string, inputArea *tview.InputField, ChatArea *tview.TextView, app *tview.Application) {

	chatStream, err := chatClient.ReceiveAllMsgs(context.Background())
	if err != nil {
		log.Fatalf("Failed to send message to server due to %v\n", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	dynamicMsg := pb.ReceiveMsgsRequest{}

	go func() {
		// Handle user input
		inputArea.SetDoneFunc(func(key tcell.Key) {
			text := inputArea.GetText()
			label := inputArea.GetLabel()
			dynamicMsg.SenderID = userID
			switch inputState {
			case stateChooseOption:
				switch text {
				case "1": // Send direct message
					dynamicMsg.MessageType = pb.MessageTypeEnum_DIRECT
					inputState = stateEnterUsername
					inputArea.SetLabel("Enter userID: ")
				case "2": // Send message to room
					dynamicMsg.MessageType = pb.MessageTypeEnum_GROUP
					inputState = stateRoomName
					inputArea.SetLabel("Enter room name to send Text: ")
				case "3": // Join/Create room
					dynamicMsg.MessageType = pb.MessageTypeEnum_GROUP
					inputState = stateRoomName
					inputArea.SetLabel("Enter room name to Create/Join: ")
				case "4": // Send message to a room
					dynamicMsg.MessageType = pb.MessageTypeEnum_ANONYMOUS
					inputState = stateEnterMessage
					inputArea.SetLabel("Enter Message: ")
				default:
					inputArea.SetLabel("(Invalid) Choose an option [1-4]: ")
				}
			case stateRoomName:
				dynamicMsg.RoomOrRecipientID = &text
				if strings.HasPrefix(label, "Enter room name to send Text:") {
					inputState = stateEnterMessage
					inputArea.SetLabel("Enter Message: ")
				} else {
					inputArea.SetLabel("Choose an option [1-4]: ")
					optionsArea.SetText("Options:\n1. Send direct message to users\n2. Send message to a room\n3. Create / Join a room\n4. Send Anonymous Global Message")
				}
			case stateEnterUsername:
				// Store the username or perform validation here
				dynamicMsg.RoomOrRecipientID = &text
				inputState = stateEnterMessage
				inputArea.SetLabel("Enter message: ")
			case stateEnterMessage:
				dynamicMsg.Message = text
				err := chatStream.Send(&dynamicMsg)
				if err != nil {
					log.Fatalf("Failed to send the chat to server %v\n", err)
				}
				// Process the message here
				// ChatArea.Write([]byte(fmt.Sprintf("Sending Message to %s: %s\n", *dynamicMsg.RoomOrRecipientID, dynamicMsg.Message)))
				dynamicMsg = pb.ReceiveMsgsRequest{}
				inputState = stateChooseOption

				inputArea.SetLabel("Choose an option [1-4]: ")
				optionsArea.SetText("Options:\n1. Send direct message to users\n2. Send message to a room\n3. Create / Join a room\n4. Send Anonymous Global Message")

			}
			inputArea.SetText("")
		})

		// time.Sleep(time.Second * 100)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for {
			stream, err := chatStream.Recv()
			if err == io.EOF {
				log.Println("end of chat server responses")
			}
			if err != nil {
				log.Fatalf("Failed to receive the chat from server %v\n", err)
			}
			// switch
			printMessage(stream, ChatArea, app)
		}
	}()
	wg.Wait()

}

func printMessage(msg *pb.ReceiveMsgsResponse, ChatArea *tview.TextView, app *tview.Application) {
	//
	msgType := msg.GetMessageType()
	switch msgType {
	case pb.MessageTypeEnum_DIRECT:
		// log.Printf("Received Direct Message: From : %v \t Message: %v \n", msg.GetSenderID(), msg.GetMessage())
		ChatArea.Write([]byte(fmt.Sprintf("Received Direct Message: From : %v \t Message: %v \n", msg.GetSenderID(), msg.GetMessage())))
		app.Draw()
	case pb.MessageTypeEnum_GROUP:
		ChatArea.Write([]byte(fmt.Sprintf("Received Message in Room (%v): From: %v \t Message:%v\n", msg.GetRoomOrRecipientID(), msg.GetSenderID(), msg.GetMessage())))
		app.Draw()
		// log.Printf("Received Message in Room (%v): From: %v \t Message:%v\n", msg.GetRoomOrRecipientID(), msg.GetSenderID(), msg.GetMessage())
	case pb.MessageTypeEnum_ANONYMOUS:
		ChatArea.Write([]byte(fmt.Sprintf("Received Anonymous Message: %v\n", msg.GetMessage())))
		app.Draw()
		// log.Printf("Received Anonymous Message: %v\n", msg.GetMessage())
	}
}
