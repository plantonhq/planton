# CloudflareZeroTrustGatewaySettings Terraform Module

Terraform IaC module for the Secure Web Gateway configuration, logging controls, and PAC files.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustGatewaySettingsSpec (generated)
main.tf       — cloudflare_zero_trust_gateway_settings (count-gated) +
                cloudflare_zero_trust_gateway_logging (count-gated) +
                cloudflare_zero_trust_gateway_pacfile (for_each by name)
outputs.tf    — account_id, pacfile_ids
```

## Behavior

Three folded lifecycles:

- **settings** — the account configuration SINGLETON: create==update (same PUT), destroy is a NO-OP abandoning the live configuration. An unset spec sub-object is never sent (not managed). Import as `{account_id}`.
- **logging** — the logging SINGLETON, same lifecycle: the COMPLETE tree is always sent when declared, because partial sends drift forever at Cloudflare. Import as `{account_id}`.
- **pac_files** — one resource per row keyed by name, real create/update/delete; `slug` forces replacement (baked into the public URL). An omitted slug is derived deterministically from the name (lowercased, non-alphanumerics to hyphens) — a server-generated slug is random and would make every refreshed plan propose recreating the file (live-measured at v5.23.0). Import as `{account_id}/{pacfile_id}`; the server-assigned file ids are exported in the `pacfile_ids` output keyed by file name, so import IDs read from a prior deploy's outputs instead of listing the API collection.

Enabling `tls_decrypt` before a Gateway certificate is activated fails at the API with error 2211.

## Outputs

| Name | Description |
|------|-------------|
| `account_id` | The account the configuration was applied to (the singleton's identity) |
| `pacfile_ids` | Cloudflare-assigned PAC-file ids keyed by file name (import recipes derive per-file import IDs from it) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
