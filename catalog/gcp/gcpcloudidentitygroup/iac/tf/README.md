# GcpCloudIdentityGroup — Terraform Implementation

This directory contains the Terraform implementation for provisioning a
Google Group in Cloud Identity from the Planton spec: one
`google_cloud_identity_group` under the customer and one
`google_cloud_identity_group_membership` per member.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration); the principal needs the Groups
  Admin role in the Cloud Identity customer

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Display name (spec or metadata.name fallback), the label map derived from `security`, the defaulted `initial_group_config`, the memberships keyed by email |
| `main.tf` | `google_cloud_identity_group` + `google_cloud_identity_group_membership` (`for_each` by member email) |
| `outputs.tf` | `name`, `group_email`, `membership_count` |

## Send Posture

- **`labels`** -- Google accepts exactly two labels, both with empty
  values: the discussion-forum label every Google Group must carry and the
  security label that makes it a security group. The map is derived from
  `spec.security` in `locals.tf`, never a user-facing map (PARITY with the
  Pulumi module).
- **`group_key`** -- the spec's `group_email` and `group_namespace`; the
  namespace is sent only when set.
- **`initial_group_config`** -- carries a proto default (EMPTY) the
  manifest loader applies; the `coalesce` is the same rule for a
  hand-written tfvars file, and the value is always sent.
- **Memberships** -- one resource per member keyed by the flattened email,
  so a member removed from the list is destroyed by key and the others are
  left alone. Unset roles mean MEMBER only; an `expire_time` rides its
  role as `expiry_detail`; `create_ignore_already_exists` is sent only
  when true (the provider default is false).
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON, fanned
  out to the memberships so they share the group's fate.
