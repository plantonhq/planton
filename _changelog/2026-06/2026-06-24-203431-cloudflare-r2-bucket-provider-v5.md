# CloudflareR2Bucket on Cloudflare provider v5

**Date**: June 24, 2026
**Type**: Bug fix / Migration
**Components**: API Definitions, IaC Modules (OpenTofu + Pulumi), Testing Framework

## Summary

`CloudflareR2Bucket` now deploys cleanly on Cloudflare provider v5 across both IaC
engines, with tofu<->pulumi parity and stack outputs that match the proto contract.
The OpenTofu module previously pinned the provider to `~> 4.0` while declaring the
v5-only `cloudflare_r2_custom_domain`, so every tofu deploy failed schema validation
with `Invalid resource type`. The module is now on `~> 5.0` (matching the Pulumi
module, already on `sdk/v6`), and the v5 attribute, enum, and output shapes are
applied consistently on both engines.

## What changed

- **Provider pin**: OpenTofu module moved to Cloudflare `~> 5.0`.
- **Custom domain**: `cloudflare_r2_custom_domain` uses the v5 attributes `domain`
  and `enabled` on both engines.
- **Location hint**: `CloudflareR2Location` enum values match the provider strings
  exactly (`auto`, `wnam`, `enam`, `weur`, `eeur`, `apac`, `oc`), so the generated
  enum is used directly. `auto` is the recommended default and means "no hint" —
  both engines omit the attribute in that case and let Cloudflare choose the region.
  `location` is optional (the provider treats it as optional/computed).
- **Tofu variable shapes**: `location` is typed `optional(string)` and the custom
  domain `zone_id` is typed `optional(string)` — both match how the manifest is
  rendered into tfvars (enums as strings; `StringValueOrRef` flattened to a string).
- **Stack outputs**: both engines emit `bucket_name`, `bucket_url`
  (`https://<account_id>.r2.cloudflarestorage.com/<bucket>`), and `custom_domain_url`
  (only when a custom domain is enabled), matching `stack_outputs.proto`.

## Testing

- `make protos`, `go build`, and `go vet` of the component are green.
- `go test` for the component spec (including location-omitted and region cases),
  `pkg/outputs` (a new `CloudflareR2Bucket` conformance case), and
  `pkg/secretcoverage` all pass; `planton secret-coverage --check` is green.
- `tofu init`/`validate`/`plan` succeed on provider v5 against the hack manifest,
  both with and without a custom domain; a real `apply`/`destroy` was exercised
  against a sandbox account.

---

**Status**: ✅ Production Ready
