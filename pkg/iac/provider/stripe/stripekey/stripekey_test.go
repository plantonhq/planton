package stripekey

import (
	"strings"
	"testing"

	stripeprovider "github.com/plantonhq/planton/catalog/stripe"
)

func TestModeOf(t *testing.T) {
	for key, want := range map[string]stripeprovider.StripeMode{
		"rk_test_abc": stripeprovider.StripeMode_test,
		"sk_test_abc": stripeprovider.StripeMode_test,
		"rk_live_abc": stripeprovider.StripeMode_live,
		"sk_live_abc": stripeprovider.StripeMode_live,
	} {
		got, prefix, ok := ModeOf(key)
		if !ok || got != want || !strings.HasPrefix(key, prefix) {
			t.Errorf("ModeOf(%q) = %v, %q, %v; want %v with its prefix", key, got, prefix, ok, want)
		}
	}
	for _, key := range []string{"pk_test_abc", "whsec_abc", "", "rk_abc"} {
		if _, _, ok := ModeOf(key); ok {
			t.Errorf("ModeOf(%q) must not recognize a non-secret, non-restricted key", key)
		}
	}
}

func TestRequireTestMode_RefusesEverythingButATestKey(t *testing.T) {
	for _, key := range []string{"rk_test_secretvalue", "sk_test_secretvalue"} {
		if err := RequireTestMode(key); err != nil {
			t.Errorf("%s...: a test key passes, got %v", key[:8], err)
		}
	}
	for _, key := range []string{"rk_live_secretvalue", "sk_live_secretvalue"} {
		err := RequireTestMode(key)
		if err == nil || !strings.Contains(err.Error(), "live-mode key ("+key[:8]+"...)") {
			t.Errorf("%s...: a live key must be refused by its prefix, got %v", key[:8], err)
		}
		if err != nil && strings.Contains(err.Error(), "secretvalue") {
			t.Error("a refusal names the prefix, never the key")
		}
	}
	if err := RequireTestMode("pk_test_secretvalue"); err == nil || strings.Contains(err.Error(), "secretvalue") {
		t.Errorf("a publishable key is refused without echoing it, got %v", err)
	}
}
