# Team Discussion List

## Use Case

A plain Google Group for a team's mail and calendar invitations -- a discussion forum, not an access-control object -- with the deploying principal made an owner so the list is never orphaned when its human owner leaves. Members without explicit roles are plain MEMBERs.

## When to Use

- Team mailing lists and calendar audiences
- Announcement lists and on-call rotations that receive mail
- Anything that is NOT an IAM binding target (use the security-group preset for those)

## What This Creates

- A plain Google Group `data-team@example.com` under customer `C01abc2de`
- `WITH_INITIAL_OWNER`: the deploying principal is added as an owner beyond the declared list
- Three memberships: one manager and two plain members
- `deletionPolicy: DELETE`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `groupEmail` | `data-team@example.com` | The list's address in a domain your customer owns. Immutable. |
| `memberships[]` | three people | Your team; unset `roles` means MEMBER only. |
| `security` | `false` | `true` only if IAM will bind the group -- the label cannot be removed later. |
| `deletionPolicy` | `DELETE` | `PREVENT` once other systems depend on the address. |

Members Google adds outside the manifest (the initial owner, Admin-console edits) are not managed here and are left alone.
