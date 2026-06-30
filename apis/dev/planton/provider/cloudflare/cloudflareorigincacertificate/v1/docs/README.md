# CloudflareOriginCaCertificate — Research & Design Notes

## Purpose

A Cloudflare Origin CA certificate (`cloudflare_origin_ca_certificate`) is a free
TLS certificate, signed by Cloudflare's own Origin CA, that is trusted by
Cloudflare's edge for connections to an origin. It enables "Full (Strict)" SSL
without buying a publicly-trusted certificate for the origin. It is not valid in
browsers — only between Cloudflare and the origin.

## Scope and the key-generation decision

The provider resource takes a CSR and returns the signed certificate; the private
key is the caller's responsibility. A certificate is useless to a downstream origin
without its key, so to make this a genuinely composable node the component
generates the key + CSR by default (via the `tls` provider — `tls_private_key` /
`tls_cert_request` in Terraform, `NewPrivateKey` / `NewCertRequest` in Pulumi) and
exports the generated private key as a sensitive output. Advanced users who already
hold a key supply `spec.csr`; then no key is generated and `private_key` is empty.

The key algorithm follows `request_type`: `origin-rsa` → RSA 2048, `origin-ecc` →
ECDSA P-256. `keyless-certificate` users supply their own CSR (the key lives on
their key server).

## Scope (no zone_id / account_id)

The resource is user/account-scoped via the API token; there is no `zone_id` or
`account_id` field. The `hostnames` may belong to any zone the token controls.

## Authentication

A standard API token with `Zone : SSL and Certificates : Edit` issues Origin CA
certificates. The legacy Origin CA Key (`X-Auth-User-Service-Key`) is deprecated
upstream and is not used here.

## Engine parity

Both engines generate the key + CSR identically when `csr` is omitted and export
the same outputs (`certificate_id`, `certificate`, `private_key`, `expires_on`).
The `certificate` output is intentionally non-sensitive (public material) on both
sides; the generated `private_key` is sensitive on both sides. No
`PARITY-EXCEPTION` is required at pulumi-cloudflare v6.17.0 / provider v5.

## Gotchas

- Changing `hostnames`, `request_type`, or `requested_validity` re-issues the
  certificate (the upstream attributes force replacement).
- The certificate is not browser-trusted; use it only for the Cloudflare-to-origin
  hop with the zone's SSL mode set to Full (Strict).
