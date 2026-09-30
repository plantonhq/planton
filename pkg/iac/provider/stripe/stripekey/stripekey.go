// Package stripekey reads the mode a Stripe API key belongs to from its prefix. Stripe's
// Terraform provider applies whatever key it is handed to that key's account and has no guard
// against live mode, so every place that must keep a key in its lane -- the credential loader
// holding a key to its config's declared mode, the live-test harness refusing any live key --
// reads the mode here, in one spelling.
package stripekey

import (
	"strings"

	"github.com/pkg/errors"
	stripeprovider "github.com/plantonhq/planton/catalog/stripe"
)

// ModeOf reads the mode from a secret (sk_) or restricted (rk_) key's prefix: sk_test_ and
// rk_test_ are test mode, which every sandbox key is; sk_live_ and rk_live_ are live. It returns
// the prefix for messages, which name the prefix and never the key, and ok false for anything
// else -- a publishable pk_ key cannot manage objects.
func ModeOf(apiKey string) (mode stripeprovider.StripeMode, prefix string, ok bool) {
	for _, kind := range []string{"sk_", "rk_"} {
		switch {
		case strings.HasPrefix(apiKey, kind+"test_"):
			return stripeprovider.StripeMode_test, kind + "test_", true
		case strings.HasPrefix(apiKey, kind+"live_"):
			return stripeprovider.StripeMode_live, kind + "live_", true
		}
	}
	return stripeprovider.StripeMode_stripe_mode_unspecified, "", false
}

// RequireTestMode refuses any key that is not a test-mode secret or restricted key. It is the
// guard for surfaces that must never touch real money whatever they are configured with, such as
// the live-test lanes, which run only against a dedicated sandbox.
func RequireTestMode(apiKey string) error {
	mode, prefix, ok := ModeOf(apiKey)
	switch {
	case !ok:
		return errors.New("the Stripe key is not a secret or restricted key: use a sandbox key that starts with rk_test_ or sk_test_")
	case mode != stripeprovider.StripeMode_test:
		return errors.Errorf("the Stripe key is a live-mode key (%s...), and this runs only against a test sandbox: use the sandbox's rk_test_ or sk_test_ key", prefix)
	}
	return nil
}
