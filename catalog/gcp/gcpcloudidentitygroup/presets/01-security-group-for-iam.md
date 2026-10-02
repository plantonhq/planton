# Security Group for IAM

## Use Case

The group every IAM binding should name instead of people: a security group with a human owner, a manager who can maintain the membership, a deployer service account, and a contractor whose membership expires on a date. Bind `roles/owner`, `roles/editor`, or custom roles to `group:platform-admins@example.com` through `GcpProjectIamMember`; offboarding becomes one membership edit.

## When to Use

- The admin, developer, or viewer group for a project, folder, or organization
- Any group IAM will bind (security groups are built for access control)
- Time-boxed access for contractors and break-glass responders

## What This Creates

- A security group `platform-admins@example.com` under customer `C01abc2de`
- Four memberships: an owner, a manager, a service account (by `GcpServiceAccount` reference), and a MEMBER expiring 2027-03-31
- `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `groupEmail` | `platform-admins@example.com` | A stable role name in a domain your customer owns. Immutable. |
| `customerId` | `customers/C01abc2de` | Your Cloud Identity customer (Admin console > Account). Immutable. |
| `memberships[]` | four members | Your people, groups, and `GcpServiceAccount` references; every listed role set includes `MEMBER`. |
| `expireTime` | one contractor | RFC 3339 timestamps on `MEMBER` roles that should lapse. |
| `initialGroupConfig` | `EMPTY` | `WITH_INITIAL_OWNER` to make the deploying principal an owner too. |

Then bind roles: a `GcpProjectIamMember` with `member: group:platform-admins@example.com`.
