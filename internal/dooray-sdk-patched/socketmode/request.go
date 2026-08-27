package socketmode

import "github.com/dooray-go/dooray-sdk/openapi/messenger"

// SocketModeRequest represents an incoming event from the Dooray WebSocket connection.
type SocketModeRequest struct {
	EnvelopeID string                 `json:"envelope_id"`
	Type       string                 `json:"type"`
	Service    string                 `json:"service"`
	Action     string                 `json:"action"`
	Payload    map[string]interface{} `json:"payload"`
	Entity     *Entity                `json:"entity"`
	Actor      *Actor                 `json:"actor"`
	ActionData *DataWrapper           `json:"actionData"`

	// agent is a back-reference used by Reply().
	agent *Agent
}

// Entity wraps the event entity data.
type Entity struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// Actor wraps information about the user who triggered the event.
type Actor struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// DataWrapper wraps additional action detail data.
type DataWrapper struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// IsMessage returns true if the event is a messenger message.
func (r *SocketModeRequest) IsMessage() bool {
	return r.Service == ServiceMessenger && r.Type == "message"
}

// Text extracts the message text from the entity data.
func (r *SocketModeRequest) Text() string {
	if r.Entity == nil || r.Entity.Data == nil {
		return ""
	}
	if text, ok := r.Entity.Data["text"].(string); ok {
		return text
	}
	return ""
}

// ChannelID extracts the channel ID from the entity data.
func (r *SocketModeRequest) ChannelID() string {
	if r.Entity == nil || r.Entity.Data == nil {
		return ""
	}
	if ch, ok := r.Entity.Data["channelId"].(string); ok {
		return ch
	}
	return ""
}

// SenderID extracts the sender member ID from the entity data.
func (r *SocketModeRequest) SenderID() string {
	if r.Entity == nil || r.Entity.Data == nil {
		return ""
	}
	if id, ok := r.Entity.Data["senderId"].(string); ok {
		return id
	}
	return ""
}

// IsBotMessage returns true if the message was sent by this bot itself.
// It compares the sender ID with the bot's own organizationMemberId
// obtained during socket mode token exchange.
// Use this to avoid infinite loops when the bot replies to its own messages.
func (r *SocketModeRequest) IsBotMessage() bool {
	if r.agent == nil || r.agent.memberID == "" {
		return false
	}
	return r.SenderID() == r.agent.memberID
}

// IsType checks if the event type matches.
func (r *SocketModeRequest) IsType(t string) bool {
	return r.Type == t
}

// IsAction checks if the event action matches.
func (r *SocketModeRequest) IsAction(a string) bool {
	return r.Action == a
}

// IsService checks if the event service matches.
func (r *SocketModeRequest) IsService(s string) bool {
	return r.Service == s
}

// Reply sends a text message back to the channel where the event originated.
// Only works for messenger events with a channel ID.
func (r *SocketModeRequest) Reply(text string) error {
	ch := r.ChannelID()
	if ch == "" {
		return ErrNoChannel
	}
	if r.agent == nil || r.agent.messenger == nil {
		return ErrNoWebClient
	}
	_, err := r.agent.messenger.SendMessage(r.agent.agentToken, ch, &messenger.SendMessageRequest{Text: text})
	return err
}
