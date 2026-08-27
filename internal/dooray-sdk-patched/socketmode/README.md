# Socket Mode

Socket Mode allows your bot to receive real-time events from Dooray via a persistent WebSocket connection, without setting up a public webhook endpoint.

## Prerequisites

- **Agent Token**: `DOORAY_AGENT_TOKEN` environment variable (issued from Dooray agent settings)
- **Domain** (optional): `DOORAY_DOMAIN` environment variable (e.g. `company`)

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/dooray-go/dooray-sdk/socketmode"
)

func main() {
    agent := socketmode.NewAgent(os.Getenv("DOORAY_AGENT_TOKEN"),
        socketmode.WithDomain(os.Getenv("DOORAY_DOMAIN")),
    )

    agent.OnMessenger(func(req *socketmode.SocketModeRequest) {
        if text := req.Text(); text != "" {
            req.Reply("Echo: " + text)
        }
    })

    if err := agent.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Agent Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithBaseURL(url)` | Dooray API base URL | `https://api.dooray.com` |
| `WithDomain(domain)` | Dooray domain name | (empty) |
| `WithHTTPClient(client)` | Custom `*http.Client` for REST API calls | `http.DefaultClient` |
| `WithLogger(logger)` | Custom `*log.Logger` | stdout with `[dooray-socketmode]` prefix |
| `WithPingInterval(d)` | WebSocket ping interval | `30s` |
| `WithReconnectBackoff(min, max)` | Reconnect backoff range | `1s` ~ `30s` |

```go
agent := socketmode.NewAgent(token,
    socketmode.WithBaseURL("https://api.dooray.com"),
    socketmode.WithDomain("mycompany"),
    socketmode.WithPingInterval(20*time.Second),
    socketmode.WithReconnectBackoff(2*time.Second, 60*time.Second),
)
```

## Event Handlers

### Service-specific handlers

Register handlers for a specific Dooray service. Each handler receives all events from that service.

```go
// Messenger events (messages, reactions, etc.)
agent.OnMessenger(func(req *socketmode.SocketModeRequest) {
    fmt.Printf("Messenger event: type=%s action=%s\n", req.Type, req.Action)
})

// Task events (create, update, status change, etc.)
agent.OnTask(func(req *socketmode.SocketModeRequest) {
    fmt.Printf("Task event: type=%s action=%s\n", req.Type, req.Action)
})

// Wiki events (page create, update, etc.)
agent.OnWiki(func(req *socketmode.SocketModeRequest) {
    fmt.Printf("Wiki event: type=%s action=%s\n", req.Type, req.Action)
})
```

### Filtered handler with `On()`

Register a handler with fine-grained filtering. Empty string matches all values for that field.

```go
// Only handle new message creation in messenger
agent.On("messenger", "message", "create", func(req *socketmode.SocketModeRequest) {
    fmt.Printf("New message: %s\n", req.Text())
})

// Handle all task events regardless of type and action
agent.On("task", "", "", func(req *socketmode.SocketModeRequest) {
    fmt.Println("Something happened in task service")
})

// Handle all "create" actions across all services
agent.On("", "", "create", func(req *socketmode.SocketModeRequest) {
    fmt.Printf("Something was created in %s\n", req.Service)
})
```

## SocketModeRequest

| Method | Return | Description |
|--------|--------|-------------|
| `Text()` | `string` | Message text from entity data |
| `ChannelID()` | `string` | Channel ID from entity data |
| `SenderID()` | `string` | Sender member ID from entity data |
| `IsMessage()` | `bool` | True if messenger message event |
| `IsType(t)` | `bool` | Check event type |
| `IsAction(a)` | `bool` | Check event action |
| `IsService(s)` | `bool` | Check event service |
| `Reply(text)` | `error` | Send a reply to the originating channel |

### Fields

```go
type SocketModeRequest struct {
    EnvelopeID string              // Unique message identifier
    Type       string              // Event type (e.g. "message")
    Service    string              // Service name ("messenger", "task", "wiki")
    Action     string              // Action type ("create", "update", "delete")
    Payload    map[string]any      // Raw event payload
    Entity     *Entity             // Event entity (type + data)
    Actor      *Actor              // Who triggered the event (type + data)
    ActionData *DataWrapper        // Additional action details
}
```

### Entity data for messenger events

```go
agent.OnMessenger(func(req *socketmode.SocketModeRequest) {
    if req.Entity != nil {
        data := req.Entity.Data
        fmt.Println("id:", data["id"])
        fmt.Println("channelId:", data["channelId"])
        fmt.Println("senderId:", data["senderId"])
        fmt.Println("text:", data["text"])
        fmt.Println("sentAt:", data["sentAt"])
    }
})
```

## Using the Messenger Client

The agent exposes the underlying `openapi/messenger.Messenger` client for direct REST API calls.

```go
agent := socketmode.NewAgent(token)
m := agent.Messenger()

// Send a message to a channel
m.SendMessage(token, "channel-id", &messenger.SendMessageRequest{
    Text: "Hello from bot!",
})

// Send a direct message to a user
m.DirectSend(token, &messenger.DirectSendRequest{
    Text:                 "Hello!",
    OrganizationMemberId: "member-id",
})
```

## Graceful Shutdown

`Run()` handles `SIGINT` and `SIGTERM` automatically. Use `RunContext()` for custom context control.

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
defer cancel()

if err := agent.RunContext(ctx); err != nil {
    log.Fatal(err)
}
```

## Connection Lifecycle

1. `POST /agent/v1/websocket/connect` to obtain a WebSocket URL
2. Connect via WebSocket with `dooray-api` authorization header
3. Receive events and auto-acknowledge with `envelope_id`
4. Ping/pong keeps the connection alive
5. On disconnect, reconnect with exponential backoff (jitter included)
