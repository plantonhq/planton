package stripe

// CLI help constants for the Stripe provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Stripe provider.
// These are read by the Terraform provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	"STRIPE_API_KEY",
	"STRIPE_ACCOUNT",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export STRIPE_API_KEY="<your-stripe-restricted-key>"`

// ConfigFileExample provides an example YAML configuration file. The mode is required: a key whose
// prefix belongs to the other mode is refused before any deploy runs.
const ConfigFileExample = `api_key: "<your-stripe-restricted-key>"
mode: test`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "stripe-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Stripe"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/stripe/stripe/latest/docs"
