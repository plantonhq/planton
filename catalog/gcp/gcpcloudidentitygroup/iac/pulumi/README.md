# GcpCloudIdentityGroup — Pulumi Implementation

This directory contains the Pulumi implementation for provisioning a
Google Group in Cloud Identity from the Planton spec: one
`gcp.cloudidentity.Group` under the customer and one
`gcp.cloudidentity.GroupMembership` per member.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `group` |
| `module/locals.go` | Display name (spec or metadata.name fallback), the label map derived from `security` |
| `module/group.go` | Maps spec to `gcp.cloudidentity.Group` and one `GroupMembership` per member; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `group_email`, `membership_count`) |

## Send Posture (parity with Terraform)

- **`Labels`** -- derived from `spec.security`: the discussion-forum label
  always, the security label on top when true; never a user-facing map.
- **`GroupKey`** -- the spec's `group_email` and `group_namespace` (the
  namespace only when set).
- **`InitialGroupConfig`** -- sent explicitly with its proto default
  (EMPTY) when the spec leaves it empty, exactly as the Terraform module's
  `coalesce`.
- **Memberships** -- one resource per member, named by email and parented
  to the group. Unset roles mean MEMBER only; an `ExpireTime` rides its
  role as `ExpiryDetail`; `CreateIgnoreAlreadyExists` only when true.
- **`DeletionPolicy`** -- DELETE (default), PREVENT, or ABANDON, fanned out
  to the memberships; sent only when set on both engines.
