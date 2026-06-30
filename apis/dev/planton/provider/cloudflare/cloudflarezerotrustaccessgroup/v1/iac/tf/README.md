# Terraform Module: Cloudflare Zero Trust Access Group

Provisions a single `cloudflare_zero_trust_access_group` — a reusable bundle of
Access rules referenced by policies and other groups.

## Layout

```
iac/tf/
├── provider.tf    # cloudflare provider ~> 5.0
├── variables.tf   # metadata + spec (rules typed as list(object(...)))
├── locals.tf      # labels + scope + rule pass-through
├── main.tf        # the cloudflare_zero_trust_access_group resource
└── outputs.tf     # group_id
```

## Inputs

A `spec` matching `CloudflareZeroTrustAccessGroupSpec`. Set exactly one of
`account_id` or `zone_id`; `include` requires at least one rule. Each access rule is
an object with exactly one variant key set; the proto field names map 1:1 to the
provider's attribute names, so the rule lists pass straight through.

## Outputs

- `group_id` — referenced by a policy's `group` rule or another group.

## Requirements

- API token with **Account → Access: Organizations, Identity Providers, and Groups
  → Edit** (account-scoped groups; the equivalent zone permission for zone scope).
