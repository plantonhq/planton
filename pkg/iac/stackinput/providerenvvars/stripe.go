package providerenvvars

import (
	"github.com/pkg/errors"
	stripeprovider "github.com/plantonhq/planton/catalog/stripe"
	"github.com/plantonhq/planton/pkg/iac/provider/stripe/stripekey"
)

// loadStripeEnvVars loads the Stripe provider config and returns environment variables, refusing a
// key whose mode disagrees with the config's declared mode.
//
// Stripe's Terraform provider has no guard against live mode: whatever key it is handed, it applies
// to that key's account. The declared mode is the only thing standing between a manifest meant for a
// sandbox and a live account, so the check runs here, where every deploy on every engine assembles
// its credentials, rather than as a proto validation rule the deploy path never evaluates. A config
// with no key emits nothing and passes: the missing credential is reported by the caller that
// requires it, and an empty config must never produce an empty variable.
func loadStripeEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(stripeprovider.StripeProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Stripe provider config")
	}

	if config.ApiKey != "" {
		if err := checkStripeKeyMode(config.ApiKey, config.Mode); err != nil {
			return nil, err
		}
	}

	envVars := map[string]string{}
	putIfSet(envVars, "STRIPE_API_KEY", config.ApiKey)
	putIfSet(envVars, "STRIPE_ACCOUNT", config.StripeAccount)
	return envVars, nil
}

// checkStripeKeyMode holds a key to the declared mode. Stripe keys carry their mode in the prefix
// (stripekey.ModeOf); the refusal names the prefix and never the key.
func checkStripeKeyMode(apiKey string, declared stripeprovider.StripeMode) error {
	keyMode, prefix, ok := stripekey.ModeOf(apiKey)
	if !ok {
		return errors.New("the Stripe provider config's api_key is not a Stripe secret or restricted key: " +
			"use a key that starts with rk_test_, rk_live_, sk_test_ or sk_live_ (a publishable pk_ key cannot manage objects)")
	}
	switch declared {
	case stripeprovider.StripeMode_stripe_mode_unspecified:
		return errors.Errorf("the Stripe provider config does not declare its mode, and its key is a %s-mode key (%s...): "+
			"set mode: %s to confirm this config is meant for %s mode", keyMode, prefix, keyMode, keyMode)
	case keyMode:
		return nil
	default:
		return errors.Errorf("this Stripe provider config is for %s mode, but its key is a %s-mode key (%s...): "+
			"use a %s-mode or sandbox key, or set mode: %s if this config is meant for %s mode",
			declared, keyMode, prefix, declared, keyMode, keyMode)
	}
}
