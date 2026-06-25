# CloudflareZeroTrustAccessPolicy — design notes

## What this component is

A first-class, reusable Cloudflare Zero Trust **Access policy**: an account-scoped
decision plus the access rules that determine who it applies to. It maps to the
provider's `cloudflare_zero_trust_access_policy` resource and is attached to
applications by reference.

## Decision + rules

- `decision`: `allow`, `deny`, `non_identity` (service-token / mTLS, no human
  identity), or `bypass` (skip Access for matching requests).
- `include` (OR), `exclude` (NOT), `require` (AND) use the same `CloudflareAccessRule`
  oneof documented on `CloudflareZeroTrustAccessGroup`.

## Governance

`approval_required` + `approval_groups`, `purpose_justification_required` +
`purpose_justification_prompt`, `isolation_required`, per-policy `mfa_config`, and
`connection_rules` (RDP clipboard constraints for infrastructure targets) cover the
governance surface of the v5 policy.

## Scope

Policies are **account-scoped** (`account_id` required, 32-hex). Zone-scoped access
is achieved by attaching the policy to a zone-scoped application — the policy itself
stays account-level and reusable.

## Composability

- `group` rules reference a `CloudflareZeroTrustAccessGroup`
  (`status.outputs.group_id`).
- `approval_groups[].email_list_uuid` and the rule-level identity-provider / list /
  service-token IDs are `StringValueOrRef`, ready to become foreign keys as those
  first-class kinds are forged.
- Output `policy_id` is referenced by `CloudflareZeroTrustAccessApplication`.

## Engine parity

Both engines provision the full surface except the `cloudflare_account_member` rule
variant, which the Pulumi Cloudflare SDK (v6.17.0) does not expose. The proto models
it and Terraform provisions it; the Pulumi module logs a warning and skips it. See
the Pulumi module README.

## Outputs

- `policy_id` — the Access policy ID.
