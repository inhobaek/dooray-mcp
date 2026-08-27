package socketmode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dooray-go/dooray-sdk/openapi/messenger"
	"github.com/dooray-go/dooray-sdk/utils"
	"github.com/gorilla/websocket"
)

// HandlerFunc is the function signature for event handlers.
type HandlerFunc func(req *SocketModeRequest)

// Agent manages a WebSocket connection to Dooray and dispatches events to handlers.
type Agent struct {
	agentToken string
	baseURL    string
	domain     string

	httpClient *http.Client
	messenger  *messenger.Messenger
	logger     *log.Logger
	memberID   string // bot's own organizationMemberId, set after token fetch
	pingInterval  time.Duration
	pingTimeout   time.Duration
	reconnectMin  time.Duration
	reconnectMax  time.Duration

	mu               sync.RWMutex
	messengerHandler HandlerFunc
	taskHandler      HandlerFunc
	wikiHandler      HandlerFunc
	genericHandlers  []genericHandler
}

type genericHandler struct {
	service string
	typ     string
	action  string
	fn      HandlerFunc
}

// NewAgent creates a new socket mode Agent.
// The agentToken is the DOORAY_AGENT_TOKEN used for authentication.
func NewAgent(agentToken string, opts ...Option) *Agent {
	a := &Agent{
		agentToken:   agentToken,
		baseURL:      defaultBaseURL,
		httpClient:   utils.NewDefaultHTTPClient(),
		logger:       log.New(os.Stdout, "[dooray-socketmode] ", log.LstdFlags),
		pingInterval: defaultPingInterval,
		pingTimeout:  defaultPingTimeout,
		reconnectMin: defaultReconnectMin,
		reconnectMax: defaultReconnectMax,
	}
	for _, opt := range opts {
		opt(a)
	}
	a.messenger = messenger.NewMessengerWithClient(a.baseURL, a.httpClient)
	return a
}

// Messenger returns the underlying Messenger client for making REST API calls.
func (a *Agent) Messenger() *messenger.Messenger {
	return a.messenger
}

// OnMessenger registers a handler for all messenger service events.
func (a *Agent) OnMessenger(fn HandlerFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messengerHandler = fn
}

// OnTask registers a handler for all task service events.
func (a *Agent) OnTask(fn HandlerFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.taskHandler = fn
}

// OnWiki registers a handler for all wiki service events.
func (a *Agent) OnWiki(fn HandlerFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.wikiHandler = fn
}

// On registers a handler with optional filtering by service, type, and action.
// Empty strings match all values for that field.
func (a *Agent) On(service, typ, action string, fn HandlerFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.genericHandlers = append(a.genericHandlers, genericHandler{
		service: service,
		typ:     typ,
		action:  action,
		fn:      fn,
	})
}

// Run starts the WebSocket connection and blocks until the context is cancelled
// or a termination signal is received.
func (a *Agent) Run() error {
	return a.RunContext(context.Background())
}

// RunContext starts the WebSocket connection with the given context.
func (a *Agent) RunContext(ctx context.Context) error {
	if a.agentToken == "" {
		return ErrNoToken
	}
	if a.domain == "" {
		return ErrNoDomain
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	backoff := a.reconnectMin

	for {
		err := a.connectAndListen(ctx)
		if ctx.Err() != nil {
			a.logger.Println("shutting down")
			return nil
		}

		a.logger.Printf("connection lost: %v, reconnecting in %s", err, backoff)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}

		// exponential backoff with jitter
		backoff = time.Duration(float64(backoff) * (1.5 + rand.Float64()*0.5))
		if backoff > a.reconnectMax {
			backoff = a.reconnectMax
		}
	}
}

// socketModeToken holds credentials obtained from the token endpoint.
type socketModeToken struct {
	AccessToken          string `json:"accessToken"`
	TenantID             string `json:"tenantId"`
	OrganizationMemberID string `json:"organizationMemberId"`
}

func (a *Agent) connectAndListen(ctx context.Context) error {
	// Step 1: fetch socket mode token
	token, err := a.fetchSocketModeToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch socket mode token: %w", err)
	}

	a.memberID = token.OrganizationMemberID

	// Step 2: build WebSocket URL from domain + token info
	wsURL := a.buildWebSocketURL(token)
	a.logger.Printf("connecting to %s", wsURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	conn, _, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	defer conn.Close()

	a.logger.Println("connected")

	done := make(chan struct{})
	defer close(done)

	// ping loop
	go func() {
		ticker := time.NewTicker(a.pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(
					websocket.PingMessage, nil,
					time.Now().Add(a.pingTimeout),
				); err != nil {
					a.logger.Printf("ping failed: %v", err)
					return
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			conn.WriteMessage(websocket.CloseMessage, closeMsg)
			return ctx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(a.pingInterval + a.pingTimeout))

		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		a.logger.Printf("recv: %s", string(message))

		var raw map[string]interface{}
		if err := json.Unmarshal(message, &raw); err != nil {
			a.logger.Printf("failed to unmarshal event: %v", err)
			continue
		}

		// skip sessionInfo messages
		if msgType, _ := raw["type"].(string); msgType == "sessionInfo" {
			a.logger.Printf("session established")
			continue
		}

		req := a.parseRawMessage(raw)
		if req == nil {
			continue
		}

		// send ack if envelope_id is present
		if req.EnvelopeID != "" {
			ack := map[string]string{"envelope_id": req.EnvelopeID}
			if ackData, err := json.Marshal(ack); err == nil {
				conn.WriteMessage(websocket.TextMessage, ackData)
			}
		}

		req.agent = a
		go a.dispatch(req)
	}
}

// fetchSocketModeToken calls POST /common/v1/socket-mode/tokens to obtain
// accessToken, tenantId, and organizationMemberId for WebSocket connection.
func (a *Agent) fetchSocketModeToken(ctx context.Context) (*socketModeToken, error) {
	url := a.baseURL + "/common/v1/socket-mode/tokens"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("dooray-api %s", a.agentToken))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %s", resp.Status)
	}

	var result struct {
		Header struct {
			IsSuccessful  bool   `json:"isSuccessful"`
			ResultCode    int    `json:"resultCode"`
			ResultMessage string `json:"resultMessage"`
		} `json:"header"`
		Result socketModeToken `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	if !result.Header.IsSuccessful {
		return nil, fmt.Errorf("token request failed: %s", result.Header.ResultMessage)
	}
	if result.Result.AccessToken == "" || result.Result.TenantID == "" || result.Result.OrganizationMemberID == "" {
		return nil, fmt.Errorf("incomplete token response: missing required fields")
	}

	return &result.Result, nil
}

// buildWebSocketURL constructs wss://{domain}/messenger/v5/ws/{tenantId}/{memberId}.
func (a *Agent) buildWebSocketURL(token *socketModeToken) string {
	host := a.domain
	// strip protocol if present
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	// strip trailing slash
	host = strings.TrimRight(host, "/")

	return fmt.Sprintf("wss://%s/messenger/v5/ws/%s/%s", host, token.TenantID, token.OrganizationMemberID)
}

// parseRawMessage converts raw WebSocket JSON into a SocketModeRequest.
// Returns nil if the message should be filtered out (e.g. system messages).
func (a *Agent) parseRawMessage(raw map[string]interface{}) *SocketModeRequest {
	msgType, _ := raw["type"].(string)
	action, _ := raw["action"].(string)

	// only pass channelLog with create/update actions to handlers
	if msgType == "channelLog" && action != "create" && action != "update" {
		return nil
	}

	content, _ := raw["content"].(map[string]interface{})
	if content == nil {
		content = map[string]interface{}{}
	}

	// filter system messages (content type == 1)
	if contentType, ok := content["type"].(float64); ok && contentType == 1 {
		return nil
	}

	// ensure channelId in content
	if _, ok := content["channelId"]; !ok {
		if chID, ok := raw["channelId"]; ok {
			content["channelId"] = chID
		}
	}

	// build actor
	var actor *Actor
	if actorRaw, ok := raw["actor"].(map[string]interface{}); ok {
		actorType, _ := actorRaw["type"].(string)
		actor = &Actor{Type: actorType, Data: actorRaw}
	} else if senderID, ok := content["senderId"].(string); ok {
		actor = &Actor{
			Type: "organizationMember",
			Data: map[string]interface{}{
				"organizationMember": map[string]interface{}{"id": senderID},
			},
		}
	}

	// map channelLog → message for user handlers
	eventType := msgType
	if msgType == "channelLog" {
		eventType = "message"
	}

	return &SocketModeRequest{
		EnvelopeID: stringVal(raw, "envelope_id"),
		Type:       eventType,
		Service:    ServiceMessenger,
		Action:     action,
		Payload:    content,
		Entity:     &Entity{Type: "message", Data: content},
		Actor:      actor,
	}
}

func stringVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func (a *Agent) dispatch(req *SocketModeRequest) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// service-specific handlers
	switch req.Service {
	case ServiceMessenger:
		if a.messengerHandler != nil {
			a.messengerHandler(req)
		}
	case ServiceTask:
		if a.taskHandler != nil {
			a.taskHandler(req)
		}
	case ServiceWiki:
		if a.wikiHandler != nil {
			a.wikiHandler(req)
		}
	}

	// generic filtered handlers
	for _, h := range a.genericHandlers {
		if h.service != "" && h.service != req.Service {
			continue
		}
		if h.typ != "" && h.typ != req.Type {
			continue
		}
		if h.action != "" && h.action != req.Action {
			continue
		}
		h.fn(req)
	}
}
