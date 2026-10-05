// Package profile loads and validates E2E profile YAML files from the planton
// repository's well-known filesystem locations.
//
// Provider profiles live at {provider}/aa_e2e/profile.yaml and describe how
// E2E tests are executed for an entire cloud provider (credential approach,
// test substrate, schedule lane, required tools).
//
// Kind profiles live at {kind}/v1/e2e/profile.yaml and describe a
// single kind's E2E readiness (tier, status, validated provisioners,
// timeout, deferred reason).
//
// Both profile types follow the KRM pattern (apiVersion + kind + metadata + spec)
// with apiVersion "qa.planton.dev/v1".
package profile
