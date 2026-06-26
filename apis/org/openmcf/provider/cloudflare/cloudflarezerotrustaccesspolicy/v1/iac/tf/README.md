# Terraform Module: Cloudflare Zero Trust Access Policy

Provisions a single `cloudflare_zero_trust_access_policy` — a reusable,
account-scoped decision plus its access rules, attached to applications by reference.

## Layout

```
iac/tf/
├── provider.tf    # cloudflare provider ~> 5.0
├── variables.tf   # metadata + spec
├── locals.tf      # labels + rule pass-through + approval/connection/mfa shaping
├── main.tf        # the cloudflare_zero_trust_access_policy resource
└── outputs.tf     # policy_id
```

## Inputs

A `spec` matching `CloudflareZeroTrustAccessPolicySpec`. Required: `account_id`,
`name`, `decision`, and at least one `include` rule. Access rules pass straight
through to the provider (proto field names match the provider's 1:1).

## Outputs

- `policy_id` — referenced by an application's policies list.

## Requirements

- API token with **Account → Access: Apps and Policies → Edit**.
