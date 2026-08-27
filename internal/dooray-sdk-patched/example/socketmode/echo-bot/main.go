package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dooray-go/dooray-sdk/socketmode"
)

func main() {
	token := os.Getenv("DOORAY_AGENT_TOKEN")
	if token == "" {
		log.Fatal("DOORAY_AGENT_TOKEN is required")
	}

	opts := []socketmode.Option{}

	if domain := os.Getenv("DOORAY_DOMAIN"); domain != "" {
		opts = append(opts, socketmode.WithDomain(domain))
	}
	if baseURL := os.Getenv("DOORAY_BASE_URL"); baseURL != "" {
		opts = append(opts, socketmode.WithBaseURL(baseURL))
	}

	agent := socketmode.NewAgent(token, opts...)

	// Handle messenger events
	agent.OnMessenger(func(req *socketmode.SocketModeRequest) {
		// Ignore bot messages to avoid infinite echo loops
		if req.IsBotMessage() {
			return
		}

		text := req.Text()
		if text == "" {
			return
		}
		fmt.Printf("Message from %s in channel %s: %s\n", req.SenderID(), req.ChannelID(), text)

		if err := req.Reply(fmt.Sprintf("Echo: %s", text)); err != nil {
			log.Printf("reply failed: %v", err)
		}
	})

	// Handle task events
	agent.OnTask(func(req *socketmode.SocketModeRequest) {
		fmt.Printf("Task event: type=%s action=%s\n", req.Type, req.Action)
	})

	// Handle specific event with On()
	agent.On("messenger", "message", "create", func(req *socketmode.SocketModeRequest) {
		fmt.Printf("New message created: %s\n", req.Text())
	})

	log.Println("Starting echo bot...")
	if err := agent.Run(); err != nil {
		log.Fatal(err)
	}
}
