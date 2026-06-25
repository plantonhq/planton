# CloudflareCustomHostname — Pulumi module

Provisions a `cloudflare_custom_hostname` from the component's stack input.

## Field-name nuance (pulumi-cloudflare SDK v6.17.0)

The Pulumi SDK names two ssl fields differently from the Terraform provider; the
module maps them and behavior is identical:

- spec `ssl.custom_cert_bundle` → `CustomCertBundles`.
- spec `ssl.settings.tls_1_3` → `Tls13`.

If a newer SDK aligns these names, update `module/custom_hostname.go` (`buildSsl`)
only — the proto already reads correctly.

## Upstream/provider parity (Enterprise-gated fields)

These ssl fields are modeled in the proto and provisioned by this module, but
Cloudflare gates them to Enterprise accounts at runtime, so on a non-Enterprise
account they can only be previewed, not applied: `custom_certificate` /
`custom_cert_bundle` (+ `custom_key`), `custom_csr_id`, a selectable
`certificate_authority`, and `wildcard`. No proto change is needed when validating
on an Enterprise account.
