package secretcoverage

import "testing"

func TestLooksSensitiveByName(t *testing.T) {
	cases := []struct {
		name string
		want bool
		why  string
	}{
		// Positives -- genuine secret-bearing names.
		{"password", true, "bare secret word"},
		{"db_password", true, "secret word in a segment"},
		{"registry_password", true, "a qualified secret word: the registry-credential field shape"},
		{"client_secret", true, "secret segment"},
		{"app_secret", true, "secret segment"},
		{"secret_access_key", true, "AWS secret access key is a real secret"},
		{"api_key", true, "apikey compound"},
		{"private_key", true, "privatekey compound"},
		{"access_token", true, "token word"},
		{"refresh_token", true, "token word"},
		{"connection_string", true, "connectionstring compound"},
		{"signing_key", true, "signingkey compound"},
		{"encryption_key", true, "encryptionkey compound"},
		{"passphrase", true, "passphrase word"},
		{"creds", true, "exact ambiguous token standing alone"},

		// Negatives -- look-alikes the denylist must suppress.
		{"public_key", false, "public flips the meaning"},
		{"ssh_public_key", false, "public anywhere suppresses"},
		{"access_key_id", false, "trailing id => identifier, not the secret"},
		{"client_secret_name", false, "trailing name => reference to a secret"},
		{"secret_arn", false, "trailing arn => resource id"},
		{"token_url", false, "trailing url => OAuth endpoint, not a token"},
		{"token_endpoint", false, "trailing endpoint => not a token value"},
		{"secret_version", false, "trailing version => metadata"},
		{"signing_algorithm", false, "trailing algorithm => not a key"},
		{"password_enabled", false, "trailing enabled => a flag, not a value"},

		// Negatives -- no token at all (bare "key" is intentionally not a token).
		{"object_key", false, "bare key is not flagged by design"},
		{"partition_key", false, "bare key is not flagged by design"},
		{"display_name", false, "no secret token"},
		{"region", false, "no secret token"},
		{"bucket_name", false, "no secret token"},
	}

	for _, tc := range cases {
		if got := LooksSensitiveByName(tc.name); got != tc.want {
			t.Errorf("LooksSensitiveByName(%q) = %v, want %v (%s)", tc.name, got, tc.want, tc.why)
		}
	}
}

func TestLooksSensitive(t *testing.T) {
	cases := []struct {
		message, field string
		want           bool
		why            string
	}{
		{"SecretEnvVar", "value", true, "a secret entry's payload is the secret"},
		{"Auth0ActionSecret", "value", true, "Secret as the last word of the message"},
		{"KubernetesSecretOpaqueData", "data", true, "Secret inside the message name"},
		{"KubernetesSecretOpaqueData", "binary_data", true, "the binary payload of a secret entry"},
		{"KubernetesSecretTlsData", "tls_key", true, "a TLS private key, by the tlskey compound"},
		{"KubernetesSecretTlsData", "tls_crt", false, "a certificate is public"},
		{"EnvVar", "value", false, "a plain variable's value is not a secret"},
		{"ExternalSecretsStoreFakeEntry", "value", false, "\"Secrets\" names a store, not a secret"},
		{"SecretEnvVar", "name", false, "the entry's name is not its payload"},
		{"Auth0ActionSecret", "client_secret", true, "a secret-bearing name counts in any message"},
	}
	for _, tc := range cases {
		if got := LooksSensitive(tc.message, tc.field); got != tc.want {
			t.Errorf("LooksSensitive(%q, %q) = %v, want %v (%s)", tc.message, tc.field, got, tc.want, tc.why)
		}
	}
}
