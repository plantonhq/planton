# Workload Permissions Boundary

This preset creates a permissions-boundary policy: the ceiling on what any
principal carrying it can ever do, regardless of what its permission policies
grant. Apply it through a role's or user's `permissionsBoundary` field (via
`valueFrom` referencing this policy's `policy_arn` output). Boundaries are how
an organization safely delegates role creation -- CI pipelines and platform
tooling can mint roles freely, and none of them can escalate past the boundary.

## When to Use

- Delegating IAM role/user creation to CI pipelines or platform tooling
- Enforcing an organization-wide ceiling on workload permissions
- Satisfying "no principal may touch IAM" compliance controls mechanically

## Key Configuration Choices

- **Allow-list of workload services** -- data, messaging, and observability
  actions workloads legitimately need; effective permissions are the
  *intersection* of this boundary and the principal's permission policies
- **Explicit Deny on identity escalation** -- `iam:*`, `organizations:*`, and
  `account:*` are denied outright; an explicit Deny wins over any Allow, so no
  attached policy can restore these
- **`/boundaries/` path** -- lets an IAM condition restrict who may pass this
  boundary using a wildcard match on `arn:aws:iam::<account>:policy/boundaries/*`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |

## Related Presets

- **01-s3-read-only** -- a shared read-only permission set for bucket consumers
