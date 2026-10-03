# CloudflareZeroTrustAccessIdentityProvider Pulumi Module

Pulumi (Go) IaC module for connecting Cloudflare Access to an identity source (Google, Okta, Azure AD, GitHub, OIDC, SAML, or one-time PIN).

## Architecture

```
main.go                      — Entrypoint loading the IaC input
module/main.go               — Resources(): provider setup, resource, outputs
module/locals.go             — Locals initialization
module/identity_provider.go  — cloudflare.ZeroTrustAccessIdentityProvider
module/outputs.go            — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: same resource, same dual-scope rule, same empty-config send for `onetimepin`, same `scim_secret` marked secret via `pulumi.ToSecret`.

## Outputs

| Name | Description |
|------|-------------|
| `identity_provider_id` | UUID referenced by Access policy rules |
| `scim_base_url` | SCIM v2.0 endpoint (present when SCIM is enabled) |
| `scim_secret` | Sensitive. Minted once; does not survive import |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
