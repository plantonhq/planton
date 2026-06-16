# Secure-by-default: secret-field annotation sweep

**Date:** 2026-06-16

**Type:** Security

## Summary

Swept the cloud-resource specs and annotated every field that holds a real
user-supplied secret value with the `(org.openmcf.shared.options.sensitive) = true`
option, so that field becomes secret-by-default: downstream it can only hold a
managed-secret reference (resolved just-in-time at deploy) and never plaintext.
Heuristic false positives (KMS key identifiers, names of secrets to create,
foreign-key references, token-location fields, file paths) were given an auditable
`sensitive_exempt_reason` instead. Measured secret-coverage rose from **3.9% to
81.7%** (covered=63, exempt=26, gap=20) across all providers.

## Motivation

The `sensitive` option and the secret-coverage analyzer shipped earlier, but only a
single pilot field (`AzureLinuxWebApp` registry password) was actually annotated, so
secure-by-default protected almost none of the real surface. This sweep turns the
annotation backlog into broad, measured enforcement reach: every annotated field is
now fail-closed against plaintext at create/update/apply, with resolution happening
just-in-time on the runner.

## What's New

- **62 secret fields annotated `sensitive`** across AWS, Azure, GCP, OCI, AliCloud,
  Scaleway, Auth0, Kubernetes, DigitalOcean, Civo, Hetzner, and OpenStack —
  database/admin/master passwords, OAuth client secrets, private keys and PEM
  material, auth tokens / PATs, registry and storage access keys, pre-shared keys,
  and application-credential secrets.
- **19 heuristic false positives exempted** with a recorded `sensitive_exempt_reason`:
  KMS key identifiers (`encryption_key`, `cloud_disk_encryption_key`,
  RocketMQ `storage_secret_key`), names of secrets to create (`secret_names`),
  foreign-key references (Cognito `pre_token_generation` → Lambda, Azure
  `application_insights_connection_string` → Application Insights, CodeBuild
  `encryption_key` → KMS, `credential` → Secrets Manager ARN), OCI API Gateway token
  **location** fields (`token_header` / `token_query_param` / `token_auth_scheme`),
  Kubernetes `image_pull_secrets` (names of existing secrets), and the Istio
  DestinationRule `private_key` (a mounted file path, not inline key material).
- **20 real secrets deliberately deferred** and tracked in
  `pkg/secretcoverage/baseline.yaml`, each with a reason: secrets nested inside
  repeated messages (the create UI needs a repeated-row secret picker first), service
  env-var secret maps (pending the reference-only env decision), secret-holder kinds
  (`KubernetesSecret`, `OciVaultSecret`), and `AzureVirtualMachine.admin_password`
  (which also offers a Key Vault foreign-key reference — a product decision).

## Implementation Details

- Each annotated/exempted spec gained `import "org/openmcf/shared/options/options.proto";`
  where it was not already importing it.
- `sensitive_exempt_reason` is read **only** by the secret-coverage tooling and has no
  effect on enforcement or just-in-time resolution; a field is secret-by-default solely
  when `sensitive` is true, so an exemption can never let a field accept plaintext.
- `pkg/secretcoverage/baseline.yaml` was reduced from 99 entries to the 20 deferred
  gaps, grouped by deferral reason. The ratcheting CI gate
  (`go test ./pkg/secretcoverage/...`) fails on any new unlisted gap or stale entry.
- Generated stubs (Go/Java/TypeScript/Python) and Gazelle `BUILD.bazel` files were
  regenerated via `make protos`.

## Migration Guide

Consumers upgrading to this release should note that once a field is annotated
`sensitive`, the control plane rejects plaintext for it and requires a
`$secret/<slug>` reference (resolved just-in-time on the runner). There is no
proactive data migration: a pre-existing resource holding plaintext in a now-sensitive
field keeps running, and is asked to switch to a reference on its next
create/update/apply (the rejection error names the field and the required shape).

## Examples

```proto
// Before
string master_password = 10 [(buf.validate.field).required = true];

// After (real secret -> secret-by-default)
string master_password = 10 [
  (buf.validate.field).required = true,
  (org.openmcf.shared.options.sensitive) = true
];

// Heuristic false positive -> documented exemption (coverage-only)
string encryption_key = 19 [(org.openmcf.shared.options.sensitive_exempt_reason) =
  "Customer-managed KMS key identifier (a reference), not secret key material."];
```

## Benefits

- Secure-by-default enforcement reach grew from one field to 63 across every provider;
  plaintext can no longer be stored in those fields.
- Every "looks sensitive but isn't" field now carries an auditable justification in the
  proto itself, readable by every consumer.
- The remaining work is explicit and measured: 20 tracked gaps trending to zero, guarded
  by CI so no new unannotated secret field can ship silently.
