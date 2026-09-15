package bootstrap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckOpenBAOHealth_Initialized_Unsealed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"initialized": true,
			"sealed":      false,
		})
	}))
	defer server.Close()

	status, err := CheckOpenBAOHealth(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Initialized {
		t.Error("expected initialized=true")
	}
	if status.Sealed {
		t.Error("expected sealed=false")
	}
}

func TestCheckOpenBAOHealth_NotInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]any{
			"initialized": false,
			"sealed":      true,
		})
	}))
	defer server.Close()

	status, err := CheckOpenBAOHealth(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Initialized {
		t.Error("expected initialized=false")
	}
	if !status.Sealed {
		t.Error("expected sealed=true")
	}
}

func TestInitializeOpenBAO_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/init" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body map[string]int
		json.NewDecoder(r.Body).Decode(&body)
		if body["secret_shares"] != 5 {
			t.Errorf("expected 5 shares, got %d", body["secret_shares"])
		}
		if body["secret_threshold"] != 3 {
			t.Errorf("expected 3 threshold, got %d", body["secret_threshold"])
		}
		for _, forbidden := range []string{"recovery_shares", "recovery_threshold"} {
			if _, ok := body[forbidden]; ok {
				t.Errorf("a built-in-seal init must not carry %s (a server under the built-in seal refuses it)", forbidden)
			}
		}

		json.NewEncoder(w).Encode(map[string]any{
			"keys":       []string{"key1", "key2", "key3", "key4", "key5"},
			"root_token": "root-token-abc",
		})
	}))
	defer server.Close()

	result, err := InitializeOpenBAO(context.Background(), server.Client(), server.URL, ShamirInit(5, 3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.UnsealKeys) != 5 || len(result.RecoveryKeys) != 0 {
		t.Errorf("expected 5 unseal keys and no recovery keys, got %d/%d", len(result.UnsealKeys), len(result.RecoveryKeys))
	}
	if result.RootToken != "root-token-abc" {
		t.Errorf("expected root-token-abc, got %s", result.RootToken)
	}
}

// Under a cloud seal the request carries ONLY the recovery pair -- the server
// refuses secret_shares/secret_threshold there ("not applicable to seal
// type") -- and the response carries recovery keys, no unseal keys.
func TestInitializeOpenBAO_RecoveryShares(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]int
		json.NewDecoder(r.Body).Decode(&body)
		for _, forbidden := range []string{"secret_shares", "secret_threshold"} {
			if _, ok := body[forbidden]; ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"errors":["parameters secret_shares,secret_threshold not applicable to seal type transit"]}`))
				return
			}
		}
		if body["recovery_shares"] != 5 || body["recovery_threshold"] != 3 {
			t.Errorf("expected recovery 5/3, got %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"keys":          []string{},
			"keys_base64":   []string{},
			"recovery_keys": []string{"r1", "r2", "r3", "r4", "r5"},
			"root_token":    "root-token-abc",
		})
	}))
	defer server.Close()

	result, err := InitializeOpenBAO(context.Background(), server.Client(), server.URL, RecoveryInit(5, 3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.RecoveryKeys) != 5 || len(result.UnsealKeys) != 0 {
		t.Errorf("expected 5 recovery keys and no unseal keys, got %d/%d", len(result.RecoveryKeys), len(result.UnsealKeys))
	}
	if result.RootToken != "root-token-abc" {
		t.Errorf("expected root-token-abc, got %s", result.RootToken)
	}
}

// The server opens itself under a cloud seal: the wait returns true the
// moment health says initialized and unsealed, false when the timeout passes
// with the vault still sealed, and never calls /sys/unseal.
func TestWaitUntilUnsealed(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/unseal" {
			t.Error("the wait must never call /sys/unseal")
		}
		calls++
		sealed := calls < 3
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"initialized": true, "sealed": sealed})
	}))
	defer server.Close()

	open, err := WaitUntilUnsealed(context.Background(), server.Client(), server.URL, 10*time.Second)
	if err != nil || !open {
		t.Fatalf("expected the vault to open on the third poll, got open=%v err=%v", open, err)
	}

	sealedForever := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"initialized": true, "sealed": true})
	}))
	defer sealedForever.Close()
	open, err = WaitUntilUnsealed(context.Background(), sealedForever.Client(), sealedForever.URL, 700*time.Millisecond)
	if err != nil || open {
		t.Fatalf("expected a still-sealed verdict after the timeout, got open=%v err=%v", open, err)
	}
}

func TestInitializeOpenBAO_AlreadyInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":["Vault is already initialized"]}`))
	}))
	defer server.Close()

	_, err := InitializeOpenBAO(context.Background(), server.Client(), server.URL, ShamirInit(5, 3))
	if err == nil {
		t.Fatal("expected error for already initialized")
	}
}

func TestUnsealOpenBAO_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/unseal" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		callCount++

		sealed := callCount < 3
		json.NewEncoder(w).Encode(map[string]any{
			"sealed":   sealed,
			"t":        3,
			"n":        5,
			"progress": callCount,
		})
	}))
	defer server.Close()

	err := UnsealOpenBAO(context.Background(), server.Client(), server.URL, []string{"k1", "k2", "k3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 unseal calls, got %d", callCount)
	}
}

func TestUnsealOpenBAO_StillSealed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"sealed": true,
		})
	}))
	defer server.Close()

	err := UnsealOpenBAO(context.Background(), server.Client(), server.URL, []string{"k1", "k2"})
	if err == nil {
		t.Fatal("expected error when still sealed")
	}
}
