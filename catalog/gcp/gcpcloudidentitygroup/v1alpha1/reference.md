# GcpCloudIdentityGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudIdentityGroupSpec creates a Google Group in Cloud Identity or
Google Workspace and manages its members: the unit every IAM binding
should point at instead of individual people, so access changes are a
membership edit, not a policy edit. The group is created under a Cloud
Identity customer (`customers/{id}`), which is the organization's
account, and gets an email address in one of the customer's domains.

Two kinds of Google Group exist: a discussion forum (the default, a
mailing list that IAM can also bind) and a security group, which adds
the security label and cannot be turned back into a plain group. Members
are folded in: each is a user, group, or service account by email with
its roles; the manifest's list is the whole managed membership.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudIdentityGroup
metadata:
  name: platform-admins
spec:
  # The group's email in a domain the Cloud Identity customer owns -- its
  # identity in IAM bindings (group:platform-admins@example.com).
  groupEmail: platform-admins@example.com
  # The customer the group is created under (Admin console > Account).
  customerId: customers/C01abc2de
  displayName: Platform Admins
  description: Owners of the platform projects; bind IAM roles to this group, never to people
  # A security group: access control only, no mailing-list features; the
  # label cannot be removed once added.
  security: true
  initialGroupConfig: EMPTY
  # The managed membership: people by email, a deployer service account by
  # reference, a contractor whose membership expires.
  memberships:
    - member:
        value: alice@example.com
      roles:
        - name: MEMBER
        - name: OWNER
    - member:
        value: bob@example.com
      roles:
        - name: MEMBER
        - name: MANAGER
    - member:
        valueFrom:
          kind: GcpServiceAccount
          name: platform-deployer
          fieldPath: status.outputs.email
    - member:
        value: contractor@example.com
      roles:
        - name: MEMBER
          expireTime: "2027-03-31T00:00:00Z"
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.groupEmail` | `string` | yes |  |  |
| `spec.customerId` | `string` | yes |  |  |
| `spec.groupNamespace` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.security` | `bool` |  |  |  |
| `spec.initialGroupConfig` | `string` |  | `EMPTY` |  |
| `spec.memberships` | `[]GcpCloudIdentityGroupMembership` |  |  |  |
| `spec.memberships[].member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.memberships[].memberNamespace` | `string` |  |  |  |
| `spec.memberships[].roles` | `[]GcpCloudIdentityGroupMembershipRole` |  |  |  |
| `spec.memberships[].roles[].name` | `string` | yes |  |  |
| `spec.memberships[].roles[].expireTime` | `string` |  |  |  |
| `spec.memberships[].createIgnoreAlreadyExists` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.groupEmail

`string` · required

The group's email address, in a domain the customer owns
(e.g. platform-admins@example.com). This is the group's identity in
IAM (`group:platform-admins@example.com`). Required. Immutable.

- rule: group_email must be an email address in a domain the Cloud Identity customer owns
- rule: {"required":true}

### spec.customerId

`string` · required

The Cloud Identity customer the group is created under, as
`customers/{id}` (the customer ID from the Google Admin console or
`gcloud organizations list`). Required. Immutable.

- rule: customer_id must be of the form customers/{id}, e.g. customers/C01abc2de
- rule: {"required":true}

### spec.groupNamespace

`string`

The namespace of an identity-mapped (non-Google) group, in the form
`identitysources/{id}`. Unset means a Google Group. Immutable.

### spec.displayName

`string`

Name shown in the Admin console and Groups UI. Defaults to
metadata.name.

### spec.description

`string`

Free-text description of the group's purpose (up to 4096 characters).

- rule: {"string":{"maxLen":"4096"}}

### spec.security

`bool`

Make this a security group: adds the
cloudidentity.googleapis.com/groups.security label on top of the
discussion-forum label every Google Group carries. Security groups are
for access control only (no mailing-list features) and the label
cannot be removed once added. Immutable in practice.

### spec.initialGroupConfig

`string` · optional (explicit presence)

Who the group starts with: EMPTY (the default; the members below are
the whole membership), WITH_INITIAL_OWNER (the caller becomes an owner
so the group is never orphaned). Immutable.

- default: `EMPTY`
- rule: initial_group_config must be EMPTY or WITH_INITIAL_OWNER

### spec.memberships

`[]GcpCloudIdentityGroupMembership`

The group's members. Each is one membership resource keyed by the
member's email; a member removed here is removed from the group, one
added is added. Roles change in place.

- rule: every membership carries the MEMBER role; add MANAGER or OWNER on top
- rule: each role appears at most once on a membership

### spec.memberships[].member

`string | valueFrom` · required

The member's email: a Google user, a Google Group, or a service
account -- a GcpServiceAccount reference or a literal address.
Immutable: a different member is a different membership.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.memberships[].memberNamespace

`string`

The namespace of an identity-mapped (non-Google) member, in the form
`identitysources/{id}`. Unset means a Google-managed identity.
Immutable.

### spec.memberships[].roles

`[]GcpCloudIdentityGroupMembershipRole`

The roles the member holds. Unset means MEMBER only; listed roles must
include MEMBER. Roles change in place.

- rule: expire_time applies only to the MEMBER role

### spec.memberships[].roles[].name

`string` · required

The role: MEMBER (belongs to the group), MANAGER (manages members),
OWNER (full control, including deleting the group).

- rule: name must be OWNER, MANAGER, or MEMBER
- rule: {"required":true}

### spec.memberships[].roles[].expireTime

`string`

When the MEMBER role expires and the membership is removed, as an RFC
3339 timestamp (e.g. 2027-01-01T00:00:00Z). Unset never expires.

- rule: expire_time must be an RFC 3339 timestamp such as 2027-01-01T00:00:00Z

### spec.memberships[].createIgnoreAlreadyExists

`bool`

Skip creating the membership when one for this member already exists
(adopt it instead of failing). Unset means fail on a duplicate.

### spec.deletionPolicy

`string`

What destroy does to the group:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the group and every membership are deleted; IAM
               bindings that named the group stop granting anything
  "PREVENT" -- destroy FAILS; protects a group IAM policies depend on
  "ABANDON" -- the group leaves management but keeps existing

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `members_unique`: each member email appears once

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudIdentityGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The group's resource name, `groups/{group_id}` -- the handle the Cloud Identity API addresses it by. |
| `status.outputs.group_email` | `string` | The group's email address -- its identity in IAM bindings (`group:{email}`). |
| `status.outputs.membership_count` | `string` | Number of memberships this manifest manages on the group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.memberships[].member` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
