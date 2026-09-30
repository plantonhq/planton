package aa_e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A lane must never run with a live key: Setup refuses it before any API call.
func TestSetup_RefusesALiveKeyBeforeAnyCall(t *testing.T) {
	for key, refusal := range map[string]string{
		"rk_live_secretvalue": "live-mode key (rk_live_...)",
		"sk_live_secretvalue": "live-mode key (sk_live_...)",
		"pk_test_secretvalue": "not a secret or restricted key",
	} {
		t.Setenv(envAPIKey, key)
		err := NewHarness().Setup(context.Background())
		// The refusal's own words, not a failed API call: a key that reached Stripe would fail
		// differently, and that is exactly what must never happen.
		if err == nil || !strings.Contains(err.Error(), refusal) {
			t.Fatalf("%s...: Setup must refuse this key before any call with %q, got %v", key[:8], refusal, err)
		}
		if strings.Contains(err.Error(), "secretvalue") {
			t.Errorf("%s...: a refusal names the prefix, never the key", key[:8])
		}
	}
}

func TestSetup_RequiresAKey(t *testing.T) {
	t.Setenv(envAPIKey, "")
	if err := NewHarness().Setup(context.Background()); err == nil || !strings.Contains(err.Error(), envAPIKey) {
		t.Fatalf("an unset key must be named, got %v", err)
	}
}

func TestClient_ReadResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The same account and version headers the pinned provider sends, so an object reads back
		// in the shape the module wrote it.
		if r.Header.Get("Authorization") != "Bearer rk_test_key" || r.Header.Get("Stripe-Context") != "acct_1" ||
			r.Header.Get("Stripe-Version") != "2026-05-27.dahlia" || r.Header.Get("Stripe-Account") != "" {
			t.Errorf("headers: Authorization=%q Stripe-Context=%q Stripe-Version=%q Stripe-Account=%q", r.Header.Get("Authorization"),
				r.Header.Get("Stripe-Context"), r.Header.Get("Stripe-Version"), r.Header.Get("Stripe-Account"))
		}
		switch r.URL.Path {
		case "/v1/webhook_endpoints/we_1":
			_, _ = w.Write([]byte(`{"id":"we_1","status":"enabled"}`))
		case "/v2/core/event_destinations/ed_1":
			_, _ = w.Write([]byte(`{"id":"ed_1","status":"enabled"}`))
		case "/v1/webhook_endpoints/we_gone":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"No such webhook endpoint"}}`))
		default:
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"message":"Permission denied. Enabling Webhook Endpoints Read"}}`))
		}
	}))
	defer server.Close()
	client := NewClient("rk_test_key", "acct_1")
	client.baseURL = server.URL

	object, exists, err := client.ReadResource("v1/webhook_endpoints/we_1")
	if err != nil || !exists || object["status"] != "enabled" {
		t.Errorf("an existing object reads back, got %v %v %v", object, exists, err)
	}
	if object, exists, err := client.ReadResource("v2/core/event_destinations/ed_1"); err != nil || !exists || object["id"] != "ed_1" {
		t.Errorf("a v2 object reads back under its own version path, got %v %v %v", object, exists, err)
	}
	if _, exists, err := client.ReadResource("v1/webhook_endpoints/we_gone"); err != nil || exists {
		t.Errorf("a 404 is an honest absence, got %v %v", exists, err)
	}
	if _, _, err := client.ReadResource("v1/webhook_endpoints/we_denied"); err == nil || !strings.Contains(err.Error(), "Enabling Webhook Endpoints Read") {
		t.Errorf("any other failure carries Stripe's message, got %v", err)
	}
}
