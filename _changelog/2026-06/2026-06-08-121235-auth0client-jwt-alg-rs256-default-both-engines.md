# Auth0Client: default id_token signing alg to RS256 in both IaC engines

**Date**: June 8, 2026
**Type**: Bug Fix
**Components**: API Definitions, Kubernetes Provider, Manifest Processing

## Summary

The `Auth0Client` spec documents `jwt_configuration.alg` as defaulting to RS256, but neither
IaC engine honored that default: an omitted `alg` was passed through as unset, so Auth0 fell
back to HS256. JWKS-verifying relying parties (e.g. NextAuth) reject an HS256 id_token, which
manifested as a real login failure. Both the OpenTofu and Pulumi modules now default `alg` to
RS256 when a `jwt_configuration` block is present but `alg` is omitted.

## Problem Statement / Motivation

`apis/org/openmcf/provider/auth0/auth0client/v1/spec.proto` documents the field default:

```proto
// alg is the algorithm used to sign the JWT.
// - "HS256": HMAC using SHA-256 (symmetric, uses client secret)
// - "RS256": RSA using SHA-256 (asymmetric, uses tenant keys)
// - "PS256": RSA-PSS using SHA-256
// Default: RS256 (recommended)
string alg = 3 [...];
```

But the modules did not encode that default:

- OpenTofu: `variables.tf` declared `alg = optional(string)` (no default), and `locals.tf`
  passed `var.spec.jwt_configuration.alg` straight through -> null -> Auth0 HS256.
- Pulumi: `client.go` only set `Alg` when non-empty (`if locals.JwtConfiguration.Alg != ""`),
  otherwise omitted it -> Auth0 HS256.

Both engines were in parity at the wrong value. On a real deployment this surfaced as
`[OAUTH_CALLBACK_ERROR] unexpected JWT alg received, expected RS256, got: HS256` at the
NextAuth callback (the IdP signup/login itself succeeds; only id_token verification fails).

### Pain Points

- The documented contract (RS256) and the actual deployed behavior (HS256) disagreed.
- Consumers had to redundantly set `alg: RS256` on every web client to work around it.

## Solution / What's New

Encode the documented default in both modules, using each engine's idiomatic mechanism:

- OpenTofu `variables.tf`: `alg = optional(string, "RS256")` -- mirrors the sibling
  `secret_encoded = optional(bool, false)` in the same `jwt_configuration` object. No
  `locals.tf` change is needed (it now receives `"RS256"` when omitted).
- Pulumi `client.go`: default `Alg` to `"RS256"` when empty before setting it on the client.

This is deliberately a **module-level** default rather than a proto `(options.default)`
annotation. The proto-option path is applied only by `internal/manifest/protodefaults`
`ApplyDefaults`, which runs in the OpenMCF CLI manifest loader (`internal/manifest/load_manifest.go`).
The orchestrated deploy path renders tfvars via `pkg/iac/tofu/generators/tfvars.go`
(`protojson` with `EmitUnpopulated:false`, which prunes unset fields) and never calls
`ApplyDefaults`, so a proto-option default would not take effect there. Putting the default in
the module guarantees it on every consumer path, and is consistent with the adjacent
`secret_encoded` convention.

## Nested-message semantics

`optional(string, "RS256")` (and the Pulumi equivalent) apply only when the `jwt_configuration`
object is present but `alg` is omitted. A client that omits `jwt_configuration` entirely still
gets no block (preserving "I don't want this" intent) -- matching `ApplyDefaults`' documented
behavior for unset nested messages.

## Cross-engine parity

Both engines now produce RS256 under the same condition (jwt_configuration present, alg
omitted), so the jwt-config parity dimension is MATCH. `alg` is not a stack output, so the
`pkg/outputs/conformance_test.go` bar is unaffected.

## Impact

- Auth0 web clients deployed without an explicit `alg` are now secure-by-default (RS256),
  fixing JWKS-based id_token verification (NextAuth and similar) out of the box.
- Behavior change for any consumer that set `jwt_configuration` but omitted `alg`: it now
  receives RS256 instead of HS256. This matches the documented/recommended default; consumers
  that intentionally want HS256 must set it explicitly.

## Related Work

- Surfaced by the GoSilver environment-setup track (Auth0 + NextAuth login). The GoSilver
  chart's explicit `alg: RS256` workaround becomes redundant once the new module tag is pinned.

---

**Status**: ✅ Production Ready
