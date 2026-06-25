# Pulumi Module: Cloudflare Zero Trust Access Policy

Provisions a single `cloudflare.ZeroTrustAccessPolicy` — a reusable, account-scoped
decision plus its access rules, attached to applications by reference.

## Layout

```
iac/pulumi/
├── main.go            # entrypoint (loads stack-input, calls module.Resources)
├── Pulumi.yaml
├── Makefile
└── module/
    ├── main.go            # Resources(): provider setup + policy()
    ├── locals.go          # stack-input references
    ├── policy.go          # the cloudflare.ZeroTrustAccessPolicy + rule mappers
    └── outputs.go         # output constant names
```

## Inputs

A `CloudflareZeroTrustAccessPolicyStackInput` (target + provider config). Required:
`account_id`, `name`, `decision`, and at least one `include` rule.

## Outputs

- `policy_id` — referenced by an application's policies list.

## Requirements

- API token with **Account → Access: Apps and Policies → Edit**.

## tofu↔pulumi parity

- The `cloudflare_account_member` access-rule variant is **not exposed by the Pulumi
  Cloudflare SDK (v6.17.0)**. The proto models it and the Terraform module
  provisions it; this Pulumi module logs a warning and skips that rule. To reach
  full parity, upgrade `pulumi-cloudflare/sdk/v6` to a version exposing
  `ZeroTrustAccessPolicyInclude/Exclude/Require.CloudflareAccountMember`, wire it in
  `policy.go`, and remove this note.
