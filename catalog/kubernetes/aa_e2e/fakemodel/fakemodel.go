// Package fakemodel is a stand-in for Anthropic's messages API that lanes
// deploying an agent runtime point the runtime at, so an agent turn completes
// end to end with no key, no network and no cost. The runtime is configured
// the way a user configures it for a model gateway (an API key plus
// ANTHROPIC_BASE_URL); only the model on the far side is fake.
//
// Every POST /v1/messages answers one plain text turn that ends the agent
// loop (stop_reason end_turn), as server-sent events when the request streams
// and as JSON otherwise, echoing the request's model so the caller's routing
// sees itself. GET /v1/models answers the key check a client may make. Any
// other path answers 404 with a body that names the path, so a runtime that
// calls something this fake does not serve fails loudly instead of hanging.
//
// The wire mirrors the fake Stigmer's own install smokes use
// (test/support/src/fake-llm-upstream.ts and llm-wire.ts in stigmer/stigmer),
// which the runner's Anthropic client is proven against.
//
// One fake serves a test process: Start binds it on every interface (a kind
// pod reaches the host through host.docker.internal on Docker Desktop and
// through the kind network's gateway on Linux, never through loopback), and
// Running hands it to the verifier in the same process, which reads the
// reply text and the request count to prove the turn reached the model.
package fakemodel

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// ReplyText is the text every turn answers with: plainly not a real model's
// words, so a reply carrying it can only have come from this fake.
const ReplyText = "This is the Planton e2e fake model's reply; no real model was called."

// APIKey is the key a lane hands the runtime beside the fake's address. It
// only has to be non-empty: runtimes treat an empty key as "no model
// configured", and the fake never checks it.
const APIKey = "sk-ant-e2e-fake"

// defaultModel answers a request that names no model.
const defaultModel = "claude-sonnet-4-6"

// Model is a running fake.
type Model struct {
	server   *http.Server
	listener net.Listener
	messages atomic.Int64
	sequence atomic.Int64
}

var (
	runningMu sync.Mutex
	running   *Model
)

// Start binds a fake on a free port on every interface and serves it until
// Close. It becomes the process's running fake (see Running).
func Start() (*Model, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, errors.Wrap(err, "binding the fake model on a free port")
	}
	m := &Model{listener: listener}
	m.server = &http.Server{Handler: m}
	go func() { _ = m.server.Serve(listener) }()

	runningMu.Lock()
	running = m
	runningMu.Unlock()
	return m, nil
}

// Running returns the process's running fake, or nil when none was started.
func Running() *Model {
	runningMu.Lock()
	defer runningMu.Unlock()
	return running
}

// Port is the port the fake listens on.
func (m *Model) Port() int {
	return m.listener.Addr().(*net.TCPAddr).Port
}

// Messages is how many POST /v1/messages calls the fake has answered.
func (m *Model) Messages() int64 {
	return m.messages.Load()
}

// Close stops the fake and clears it as the running one.
func (m *Model) Close() error {
	runningMu.Lock()
	if running == m {
		running = nil
	}
	runningMu.Unlock()
	return m.server.Close()
}

// ServeHTTP answers the three shapes described in the package comment.
func (m *Model) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/models":
		writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": []any{}})
	case r.Method == http.MethodPost && r.URL.Path == "/v1/messages":
		m.answerMessages(w, r)
	default:
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": fmt.Sprintf("the Planton e2e fake model serves only POST /v1/messages and GET /v1/models; it was asked for %s %s", r.Method, r.URL.Path),
		})
	}
}

type messagesRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

func (m *Model) answerMessages(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, anthropicError("invalid_request_error", "the fake model could not read the request body: "+err.Error()))
		return
	}
	var req messagesRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, anthropicError("invalid_request_error", "the fake model could not parse the request body as JSON: "+err.Error()))
		return
	}
	m.messages.Add(1)
	model := req.Model
	if model == "" {
		model = defaultModel
	}
	id := fmt.Sprintf("msg_e2e_fake_%d", m.sequence.Add(1))
	if !req.Stream {
		writeJSON(w, http.StatusOK, textMessage(id, model))
		return
	}
	writeStream(w, id, model)
}

// textMessage is a complete text turn, the non-streaming answer. Each turn
// carries an id of its own, as a provider's do: a runtime that reduces
// messages by id forgets a turn whose id it already holds.
func textMessage(id, model string) map[string]any {
	return map[string]any{
		"id":            id,
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"content":       []any{map[string]any{"type": "text", "text": ReplyText}},
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"usage":         map[string]any{"input_tokens": 10, "output_tokens": 5},
	}
}

// writeStream expands the text turn into the event sequence Anthropic's
// streaming API sends: message_start, one text block's start, delta and
// stop, message_delta with the stop reason, message_stop.
func writeStream(w http.ResponseWriter, id, model string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	events := []struct {
		name string
		data map[string]any
	}{
		{"message_start", map[string]any{
			"type": "message_start",
			"message": map[string]any{
				"id": id, "type": "message", "role": "assistant", "content": []any{}, "model": model,
				"stop_reason": nil, "stop_sequence": nil,
				"usage": map[string]any{"input_tokens": 10, "output_tokens": 0},
			},
		}},
		{"content_block_start", map[string]any{"type": "content_block_start", "index": 0, "content_block": map[string]any{"type": "text", "text": ""}}},
		{"content_block_delta", map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "text_delta", "text": ReplyText}}},
		{"content_block_stop", map[string]any{"type": "content_block_stop", "index": 0}},
		{"message_delta", map[string]any{
			"type":  "message_delta",
			"delta": map[string]any{"stop_reason": "end_turn", "stop_sequence": nil},
			"usage": map[string]any{"output_tokens": 5},
		}},
		{"message_stop", map[string]any{"type": "message_stop"}},
	}
	flusher, _ := w.(http.Flusher)
	for _, event := range events {
		data, _ := json.Marshal(event.data)
		_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.name, data)
		if flusher != nil {
			flusher.Flush()
		}
	}
}

func anthropicError(kind, message string) map[string]any {
	return map[string]any{"type": "error", "error": map[string]any{"type": kind, "message": message}}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
