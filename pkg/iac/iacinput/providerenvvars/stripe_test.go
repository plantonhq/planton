package providerenvvars

import (
	"testing"

	stripeprovider "github.com/plantonhq/planton/catalog/stripe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func stripeConfig(t *testing.T, cfg *stripeprovider.StripeProviderConfig) []byte {
	t.Helper()
	raw, err := protojson.Marshal(cfg)
	require.NoError(t, err)
	return raw
}

func TestLoadStripeEnvVars_KeyMatchingItsMode_Emitted(t *testing.T) {
	for _, tc := range []struct {
		key  string
		mode stripeprovider.StripeMode
	}{
		{"rk_test_abc", stripeprovider.StripeMode_test},
		{"sk_test_abc", stripeprovider.StripeMode_test},
		{"rk_live_abc", stripeprovider.StripeMode_live},
		{"sk_live_abc", stripeprovider.StripeMode_live},
	} {
		t.Run(tc.key, func(t *testing.T) {
			env, err := loadStripeEnvVars(stripeConfig(t, &stripeprovider.StripeProviderConfig{
				ApiKey: tc.key, Mode: tc.mode, StripeAccount: "acct_123",
			}))
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"STRIPE_API_KEY": tc.key, "STRIPE_ACCOUNT": "acct_123"}, env)
		})
	}
}

func TestLoadStripeEnvVars_KeyOfTheOtherMode_Refused(t *testing.T) {
	for _, tc := range []struct {
		key      string
		mode     stripeprovider.StripeMode
		sentence string
	}{
		{"rk_live_secretvalue", stripeprovider.StripeMode_test, "this Stripe provider config is for test mode, but its key is a live-mode key (rk_live_...)"},
		{"sk_live_secretvalue", stripeprovider.StripeMode_test, "this Stripe provider config is for test mode, but its key is a live-mode key (sk_live_...)"},
		{"rk_test_secretvalue", stripeprovider.StripeMode_live, "this Stripe provider config is for live mode, but its key is a test-mode key (rk_test_...)"},
		{"sk_test_secretvalue", stripeprovider.StripeMode_live, "this Stripe provider config is for live mode, but its key is a test-mode key (sk_test_...)"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			env, err := loadStripeEnvVars(stripeConfig(t, &stripeprovider.StripeProviderConfig{ApiKey: tc.key, Mode: tc.mode}))
			require.Error(t, err)
			assert.Nil(t, env)
			assert.Contains(t, err.Error(), tc.sentence)
			assert.NotContains(t, err.Error(), "secretvalue", "a refusal names the key's prefix, never the key")
		})
	}
}

func TestLoadStripeEnvVars_KeyWithoutDeclaredMode_Refused(t *testing.T) {
	_, err := loadStripeEnvVars(stripeConfig(t, &stripeprovider.StripeProviderConfig{ApiKey: "rk_live_secretvalue"}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not declare its mode, and its key is a live-mode key (rk_live_...): set mode: live")
	assert.NotContains(t, err.Error(), "secretvalue")
}

func TestLoadStripeEnvVars_NotASecretOrRestrictedKey_Refused(t *testing.T) {
	_, err := loadStripeEnvVars(stripeConfig(t, &stripeprovider.StripeProviderConfig{ApiKey: "pk_test_secretvalue", Mode: stripeprovider.StripeMode_test}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a Stripe secret or restricted key")
	assert.NotContains(t, err.Error(), "secretvalue")
}
