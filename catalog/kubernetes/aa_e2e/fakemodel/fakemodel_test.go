package fakemodel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func startForTest(t *testing.T) (*Model, string) {
	t.Helper()
	m, err := Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m, fmt.Sprintf("http://127.0.0.1:%d", m.Port())
}

func post(t *testing.T, url, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestJSONTurnEndsTheLoopWithTheReplyAndEchoesTheModel(t *testing.T) {
	m, base := startForTest(t)
	resp := post(t, base+"/v1/messages", `{"model":"claude-opus-4-1","messages":[{"role":"user","content":"hi"}]}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var got struct {
		ID         string `json:"id"`
		Model      string `json:"model"`
		StopReason string `json:"stop_reason"`
		Content    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if got.Model != "claude-opus-4-1" || got.StopReason != "end_turn" {
		t.Fatalf("model %q stop_reason %q, want the request's model and end_turn", got.Model, got.StopReason)
	}
	if len(got.Content) != 1 || got.Content[0].Type != "text" || got.Content[0].Text != ReplyText {
		t.Fatalf("content %+v, want one text block carrying ReplyText", got.Content)
	}
	if m.Messages() != 1 {
		t.Fatalf("Messages() = %d, want 1", m.Messages())
	}
}

func TestStreamingTurnSendsAnthropicsEventSequence(t *testing.T) {
	_, base := startForTest(t)
	resp := post(t, base+"/v1/messages", `{"model":"claude-sonnet-4-6","stream":true,"messages":[]}`)
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	var names []string
	var text string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			names = append(names, strings.TrimPrefix(line, "event: "))
		case strings.HasPrefix(line, "data: ") && strings.Contains(line, "text_delta"):
			var delta struct {
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &delta); err != nil {
				t.Fatalf("decoding the delta: %v", err)
			}
			text += delta.Delta.Text
		}
	}
	want := "message_start content_block_start content_block_delta content_block_stop message_delta message_stop"
	if strings.Join(names, " ") != want {
		t.Fatalf("events %v, want %s", names, want)
	}
	if text != ReplyText {
		t.Fatalf("streamed text %q, want ReplyText", text)
	}
}

func TestEachTurnHasItsOwnID(t *testing.T) {
	_, base := startForTest(t)
	ids := map[string]bool{}
	for i := 0; i < 3; i++ {
		resp := post(t, base+"/v1/messages", `{"messages":[]}`)
		var got struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&got)
		ids[got.ID] = true
	}
	if len(ids) != 3 {
		t.Fatalf("ids %v, want three distinct", ids)
	}
}

func TestKeyCheckAnswersAndUnknownPathsAreLoud(t *testing.T) {
	m, base := startForTest(t)
	resp, err := http.Get(base + "/v1/models")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/models: %v %v", resp, err)
	}
	_ = resp.Body.Close()
	unknown := post(t, base+"/v1/messages/count_tokens", `{}`)
	body, _ := io.ReadAll(unknown.Body)
	if unknown.StatusCode != http.StatusNotFound || !strings.Contains(string(body), "/v1/messages/count_tokens") {
		t.Fatalf("unknown path: %d %s, want 404 naming the path", unknown.StatusCode, body)
	}
	if m.Messages() != 0 {
		t.Fatalf("Messages() = %d after no turn, want 0", m.Messages())
	}
}

func TestRunningIsTheLastStartedUntilClosed(t *testing.T) {
	m, _ := startForTest(t)
	if Running() != m {
		t.Fatal("Running() is not the fake just started")
	}
	_ = m.Close()
	if Running() != nil {
		t.Fatal("Running() still returns a closed fake")
	}
}
