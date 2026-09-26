// Name-based heuristic for the secret-coverage report: does a field's NAME look
// like it should hold a secret value?
//
// The heuristic is deliberately HIGH-PRECISION, not high-recall. It exists to
// catch the obvious cases ("password", "client_secret", "api_key") so the CI
// guardrail can fail when one ships unannotated. Recall gaps are fine: a genuine
// secret the heuristic misses is simply annotated by hand during the sweep. False
// positives are the expensive failure mode (they nag every author), so a small
// denylist filters the common look-alikes, and the proto `sensitive_exempt_reason`
// escape hatch covers the residue with an auditable justification.
//
// Runtime descriptors carry no doc comments, so the heuristic reads names only: the
// field's own name, and -- for the few generic names a secret entry uses for its payload
// ("value", "data") -- the name of the message that declares it (see LooksSensitive).
package secretcoverage

import (
	"strings"
	"unicode"
)

// compoundTokens span snake_case word boundaries, so they are matched against the
// field name with underscores removed (e.g. "api_key" -> "apikey"). Only tokens
// NOT already implied by wordTokens are listed -- "client_secret"/"auth_token" are
// caught by "secret"/"token", so they are intentionally absent to avoid redundancy.
var compoundTokens = []string{
	"apikey",
	"privatekey",
	"connectionstring",
	"signingkey",
	"encryptionkey",
	"tlskey",
}

// wordTokens match when contained in any snake_case segment ("client_secret" has
// the segment "secret"; "passwords" contains "password").
var wordTokens = []string{
	"password",
	"passwd",
	"passphrase",
	"secret",
	"token",
	"credential", // also matches "credentials"
}

// exactTokens are short, ambiguous words that are only treated as secret-ish when
// they stand alone as a whole segment, to avoid matching inside unrelated words.
var exactTokens = map[string]bool{
	"creds": true,
	"pin":   true,
	"otp":   true,
	"dsn":   true,
	"sas":   true,
}

// trailingDenylist suppresses a positive match when it is the LAST segment: these
// turn a secret-ish name into a reference/metadata about a secret rather than the
// secret value itself (e.g. "client_secret_name", "secret_arn", "token_url").
var trailingDenylist = map[string]bool{
	"id":          true,
	"name":        true,
	"arn":         true,
	"ref":         true,
	"uri":         true,
	"url":         true,
	"path":        true,
	"version":     true,
	"algorithm":   true,
	"type":        true,
	"format":      true,
	"count":       true,
	"enabled":     true,
	"disabled":    true,
	"ttl":         true,
	"expiry":      true,
	"rotation":    true,
	"fingerprint": true,
	"length":      true,
	"mode":        true,
	"policy":      true,
	"status":      true,
	"region":      true,
	"endpoint":    true,
	"host":        true,
	"port":        true,
	"namespace":   true,
	"prefix":      true,
	"suffix":      true,
}

// segmentDenylist suppresses a positive match when present ANYWHERE: "public"
// flips the meaning (a public key/token is not a secret).
var segmentDenylist = map[string]bool{
	"public": true,
}

// secretEntryPayloadFields are the generic names a secret entry gives the field that
// carries its payload. On their own they say nothing ("value" is every scalar's name), so
// they count only inside a message whose name declares it a secret (see LooksSensitive).
var secretEntryPayloadFields = map[string]bool{
	"value":       true,
	"data":        true,
	"binary_data": true,
}

// LooksSensitive reports whether a field looks like it holds a secret value, reading its
// name in the context of the message that declares it. A secret-bearing name is enough
// (LooksSensitiveByName); beyond that, the payload field of a secret entry -- a "value",
// "data" or "binary_data" field of a message whose name carries the word "Secret"
// (SecretEnvVar, Auth0ActionSecret, KubernetesSecretOpaqueData) -- is the secret itself.
// The rule codifies the catalog's own convention: every such field holds secret material,
// and a message that merely describes one (a rotation schedule's "value") takes the
// `sensitive_exempt_reason` hatch. The word must stand whole: "Secrets" (a secret store's
// name, as in ExternalSecretsStore) does not count.
func LooksSensitive(messageName, fieldName string) bool {
	if LooksSensitiveByName(fieldName) {
		return true
	}
	return secretEntryPayloadFields[strings.ToLower(fieldName)] && namesASecret(messageName)
}

// namesASecret reports whether a CamelCase message name carries "Secret" as a whole word.
func namesASecret(messageName string) bool {
	for _, word := range camelWords(messageName) {
		if word == "Secret" {
			return true
		}
	}
	return false
}

// camelWords splits a CamelCase name at each upper-case letter that starts a new word
// ("Auth0ActionSecret" -> Auth0, Action, Secret).
func camelWords(name string) []string {
	var words []string
	start := 0
	for i, r := range name {
		if i > start && unicode.IsUpper(r) {
			words = append(words, name[start:i])
			start = i
		}
	}
	if start < len(name) {
		words = append(words, name[start:])
	}
	return words
}

// LooksSensitiveByName reports whether a proto field name matches the
// high-precision secret heuristic. The input is expected to be a proto field name
// (snake_case); it is lowercased defensively.
func LooksSensitiveByName(fieldName string) bool {
	name := strings.ToLower(fieldName)
	segments := strings.Split(name, "_")

	for _, seg := range segments {
		if segmentDenylist[seg] {
			return false
		}
	}
	if len(segments) > 0 && trailingDenylist[segments[len(segments)-1]] {
		return false
	}

	joined := strings.ReplaceAll(name, "_", "")
	for _, t := range compoundTokens {
		if strings.Contains(joined, t) {
			return true
		}
	}
	for _, seg := range segments {
		for _, t := range wordTokens {
			if strings.Contains(seg, t) {
				return true
			}
		}
		if exactTokens[seg] {
			return true
		}
	}
	return false
}
