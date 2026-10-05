# Pulumi Module: Cloudflare Zero Trust Access Group

Provisions a single `cloudflare.ZeroTrustAccessGroup` — a reusable bundle of Access
rules referenced by policies and other groups.

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads iac-input, calls module.Resources)
├── Pulumi.yaml
├── Makefile
└── module/
    ├── main.go            # Resources(): provider setup + group()
    ├── locals.go          # iac-input references
    ├── group.go           # the cloudflare.ZeroTrustAccessGroup + rule mappers
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareZeroTrustAccessGroupIacInput` (target + provider config). Set exactly
one of `account_id` or `zone_id`; `include` requires at least one rule.

## Outputs

- `group_id` — referenced by a policy's `group` rule or another group.

## Requirements

- API token with **Account → Access: Organizations, Identity Providers, and Groups
  → Edit** (account-scoped groups; the equivalent zone permission for zone scope).

## tofu↔pulumi parity

- The `cloudflare_account_member` access-rule variant is **not exposed by the Pulumi
  Cloudflare SDK (v6.17.0)**. The proto models it and the Terraform module
  provisions it; this Pulumi module logs a warning and skips that rule. To reach
  full parity, upgrade `pulumi-cloudflare/sdk/v6` to a version that exposes
  `ZeroTrustAccessGroupInclude/Exclude/Require.CloudflareAccountMember`, wire it in
  `group.go`, and remove this note.
