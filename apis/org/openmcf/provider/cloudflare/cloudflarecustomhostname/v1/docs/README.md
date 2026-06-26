# CloudflareCustomHostname — Research & Design Notes

## Purpose

A custom hostname (`cloudflare_custom_hostname`) is the core of Cloudflare for SaaS:
it attaches a customer-owned domain to a SaaS provider's zone, with a per-customer
certificate that Cloudflare provisions and auto-renews. It lets a SaaS platform
extend its edge (TLS, caching, WAF) onto white-label/vanity domains.

## Scope and composition

The hostname is zone-scoped (`zone_id` → CloudflareDnsZone). It depends on the zone
having a fallback origin (CloudflareCustomHostnameFallbackOrigin) — a zone-level
singleton that is its own kind, since it is shared by every custom hostname in the
zone and cannot belong to any single hostname. Express that prerequisite as a
`metadata.relationships` `depends_on` edge in infra charts.

`custom_origin_server` is a backend endpoint and is modeled as a StringValueOrRef
(no default_kind) — the same pattern as a load balancer pool's origin `address` — so
a custom hostname can route to another resource's output. The customer's `hostname`
is their external domain and is a plain string (author-specified DNS naming, not a
produced handle).

## Field-name nuance (Pulumi SDK)

The Pulumi SDK names two fields differently from the Terraform provider; both carry
identical data and the module maps them:

- spec `ssl.custom_cert_bundle` → Pulumi `CustomCertBundles` (Terraform
  `custom_cert_bundle`).
- spec `ssl.settings.tls_1_3` → Pulumi `Tls13` (Terraform `tls_1_3`).

If a future Pulumi SDK aligns the names, only the module mapping changes — the spec
already reads correctly.

## Upstream/provider parity (Enterprise-gated fields)

Several `ssl` fields are modeled in the proto (the cloud's real capability) but are
**Enterprise-only** at runtime, so they can only be `tofu plan`/`pulumi preview`
validated on a non-Enterprise account: uploaded `custom_certificate` /
`custom_cert_bundle` (+ their `custom_key`), `custom_csr_id`, a selectable
`certificate_authority`, and `wildcard`. The spec leads; a future agent with an
Enterprise account can live-validate these without any proto change. See the module
READMEs for the same note.

## Outputs

The ownership-verification records (`ownership_verification{name,type,value}` and
`ownership_verification_http{http_url,http_body}`) are flattened to scalar outputs
(the established output grain — no nested-message outputs in this provider) so a
downstream CloudflareDnsRecord on the customer's zone can publish them.

## Engine parity

Both engines create the same custom hostname and emit the same outputs. ssl defaults
(`bundle_method` "ubiquitous", `type` "dv") are coalesced identically; unset
optionals are omitted on both sides. No `PARITY-EXCEPTION` is required at
pulumi-cloudflare v6.17.0 / provider v5 (the field-name differences above are
internal SDK naming, handled in the module, not behavioral divergences).

## Gotchas

- The hostname only activates once the customer adds the CNAME and the ownership
  TXT record (or serves the HTTP token).
- A same-account customer hostname may be rejected if the customer domain is itself
  a zone proxied in the same account; for real SaaS the customer domain is external.
