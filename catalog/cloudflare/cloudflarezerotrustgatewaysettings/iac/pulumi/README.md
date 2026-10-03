# CloudflareZeroTrustGatewaySettings Pulumi Module

Pulumi (Go) IaC module for the Secure Web Gateway configuration, logging controls, and PAC files.

## Architecture

```
main.go                      — Entrypoint loading the IaC input
module/main.go               — Resources(): provider setup, resources, outputs
module/locals.go             — Locals initialization
module/gateway_settings.go   — cloudflare.ZeroTrustGatewaySettings +
                               cloudflare.ZeroTrustGatewayLogging +
                               cloudflare.ZeroTrustGatewayPacfile (per row)
module/outputs.go            — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly, folding three lifecycles:

- **settings** — the account configuration SINGLETON: create==update (same PUT), destroy is a NO-OP abandoning the live configuration. An unset spec sub-object is never sent (not managed).
- **logging** — the logging SINGLETON: the COMPLETE tree is always sent when declared (partial sends drift forever at Cloudflare — the provider's own tests accept non-empty plans on them).
- **pac_files** — one provider resource per row, keyed by name, each with a real create/update/delete lifecycle. An omitted slug is derived deterministically from the name (lowercased, non-alphanumerics to hyphens) — a server-generated slug is random and replace-forcing, which would recreate the file on every refreshed plan (live-measured at v5.23.0).

Enabling `tls_decrypt` before a Gateway certificate is activated fails at the API with error 2211.

## Outputs

| Name | Description |
|------|-------------|
| `account_id` | The account the configuration was applied to (the singleton's identity) |
| `pacfile_ids` | Cloudflare-assigned PAC-file ids keyed by file name (import recipes derive per-file import IDs from it) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
