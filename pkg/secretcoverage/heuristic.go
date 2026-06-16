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
// Runtime descriptors carry no doc comments, so this is name-only by necessity.
package secretcoverage

import "strings"

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
