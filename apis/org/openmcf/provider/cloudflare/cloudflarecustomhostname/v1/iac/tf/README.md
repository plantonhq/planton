# CloudflareCustomHostname — Terraform module

Provisions a `cloudflare_custom_hostname` from the component's stack input.

## Upstream/provider parity (Enterprise-gated fields)

These ssl fields are modeled in the proto and provisioned by this module, but
Cloudflare gates them to Enterprise accounts at runtime, so on a non-Enterprise
account they can only be `tofu plan`-validated, not applied: `custom_certificate` /
`custom_cert_bundle` (+ `custom_key`), `custom_csr_id`, a selectable
`certificate_authority`, and `wildcard`. No proto change is needed when validating
on an Enterprise account.
