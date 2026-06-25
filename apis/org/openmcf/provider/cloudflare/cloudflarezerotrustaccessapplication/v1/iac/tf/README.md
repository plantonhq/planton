# Terraform Module: Cloudflare Zero Trust Access Application

Provisions a single `cloudflare_zero_trust_access_application`, wiring it to
standalone Access policies by reference, and exports its outputs (including `aud`).

## Layout

```
iac/tf/
├── provider.tf    # cloudflare provider ~> 5.0
├── variables.tf   # metadata + spec
├── locals.tf      # labels + type default + policy/target-criteria shaping
├── main.tf        # the cloudflare_zero_trust_access_application resource
└── outputs.tf     # application_id, aud, domain, saas_*
```

## Inputs

A `spec` matching `CloudflareZeroTrustAccessApplicationSpec`. Set exactly one of
`account_id` or `zone_id`; `domain` is required for self_hosted/ssh/vnc/rdp.
`policies[].policy` carries the referenced policy ID (mapped to the provider's `id`),
and `target_criteria.target_attributes` is rebuilt into the provider's
`{name => values}` map.

## Outputs

- `application_id`, `aud`, `domain`
- `saas_client_id`, `saas_client_secret` (sensitive), `saas_public_key`,
  `saas_sso_endpoint`, `saas_idp_entity_id`

## Requirements

- API token with **Account → Access: Apps and Policies → Edit** (and the matching
  zone permission for zone-scoped applications).
