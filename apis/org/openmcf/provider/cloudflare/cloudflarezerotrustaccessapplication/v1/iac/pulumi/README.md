# Pulumi Module: Cloudflare Zero Trust Access Application

Provisions a single `cloudflare.ZeroTrustAccessApplication`, wiring it to standalone
Access policies by reference, and exports its outputs (including the `aud` tag).

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads stack-input, calls module.Resources)
├── Pulumi.yaml
├── Makefile
└── module/
    ├── main.go            # Resources(): provider setup + application()
    ├── locals.go          # stack-input references
    ├── application.go     # the cloudflare.ZeroTrustAccessApplication + saas/scim builders
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareZeroTrustAccessApplicationStackInput` (target + provider config). Set
exactly one of `account_id` or `zone_id`; `domain` is required for
self_hosted/ssh/vnc/rdp types. `policies[].policy` references a
`CloudflareZeroTrustAccessPolicy` by ID.

## Outputs

- `application_id`, `aud`, `domain`
- `saas_client_id`, `saas_client_secret`, `saas_public_key`, `saas_sso_endpoint`,
  `saas_idp_entity_id` (SaaS apps)

## Requirements

- API token with **Account → Access: Apps and Policies → Edit** (and the matching
  zone permission for zone-scoped applications).
