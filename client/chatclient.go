package main

import (
	"fmt"
	"log"
	"strings"

	pb "github.com/grpc-chat/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rivo/tview"
)

var chatServerAddress = "localhost:50051"

var ChatArea = tview.NewTextView()
var app = tview.NewApplication()
var inputArea = tview.NewInputField()
var optionsArea = tview.NewTextView()

func main() {
	var userID string
	fmt.Print("Enter your username: ")
	_, err := fmt.Scanln(&userID)
	if err != nil {
		fmt.Printf("Error reading username: %v\n", err)
	}
	con, err := grpc.Dial(chatServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server due to %v\n", err)
	}
	defer con.Close()

	chatClientCon := pb.NewGrpcChatClient(con)

	// app := tview.NewApplication()

	// Chat message display area
	// ChatArea := tview.NewTextView()
	ChatArea.SetScrollable(true)
	ChatArea.SetDynamicColors(true)
	ChatArea.SetRegions(true)
	ChatArea.SetWordWrap(true)
	ChatArea.SetTitle("Chat Messages --" + userID)
	ChatArea.SetBorder(true)

	// Initialize ChatArea with some static chat messages
	staticMessages := []string{
		"Welcome to real time chat application using gRPC.",
		"",
	}
	ChatArea.SetText(strings.Join(staticMessages, "\n"))

	optionsArea.SetDynamicColors(true)
	optionsArea.SetRegions(true)
	optionsArea.SetWordWrap(true)
	optionsArea.SetTitle("Options")
	optionsArea.SetBorder(true)

	// Initialize optionsArea with options
	options := "1. Send direct message to users\n2. Send message to a room\n3. Create / Join a room\n4. Send Anonymous Global Message"
	optionsArea.SetText("Options:\n" + options)

	// User input area

	inputArea.SetLabel("Choose an option [1-4]: ")
	inputArea.SetFieldWidth(20)
	inputArea.SetBorder(true)

	// Layout for options and input
	optionsLayout := tview.NewFlex().SetDirection(tview.FlexRow)
	optionsLayout.AddItem(optionsArea, 7, 1, false)
	optionsLayout.AddItem(inputArea, 3, 1, true)

	// Main layout
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.AddItem(ChatArea, 0, 2, false)
	flex.AddItem(optionsLayout, 0, 1, true)
	// Set focus to inputArea
	app.SetFocus(inputArea)

	// Set root and run the application
	go connectToChatServer(chatClientCon, userID, inputArea, ChatArea, app)
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

}
