# GCP Cloud Identity Group

Creates a Google Group in Cloud Identity or Google Workspace and manages its members — the unit every IAM binding should point at instead of individual people, so access changes are a membership edit, not a policy edit. The group is created under your Cloud Identity customer and gets an email address in one of your domains; members are users, service accounts, or other groups by email, each with its roles. Mark it a security group for access control only.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Group** -- the `cloud_identity_group` under `customerId` with the email `groupEmail`, its display name and description, the discussion-forum label every Google Group carries (plus the security label when `security` is true), and its initial configuration
- **Memberships** -- one `cloud_identity_group_membership` per entry in `memberships`, keyed by the member's email, with its roles and optional expiry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose principal holds the Groups Admin role in the Cloud Identity / Workspace customer (a project IAM role does not reach groups).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Cloud Identity

- **The customer ID** -- `customers/{id}` from the Google Admin console (Account > Account settings) or `gcloud organizations list`.
- **A verified domain** for the group's email.
- **Members** -- existing Google users, groups, or service accounts (`GcpServiceAccount` references resolve to their email).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudIdentityGroup
metadata:
  name: platform-admins
spec:
  groupEmail: platform-admins@example.com
  customerId: customers/C01abc2de
  security: true
  memberships:
    - member:
        value: alice@example.com
      roles:
        - name: MEMBER
        - name: OWNER
    - member:
        valueFrom:
          kind: GcpServiceAccount
          name: platform-deployer
          fieldPath: status.outputs.email
```

```shell
planton apply -f cloud-identity-group.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `groupEmail` | `string` | The group's email in a domain the customer owns -- its IAM identity (`group:{email}`). Immutable. |
| `customerId` | `string` | `customers/{id}`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Name in the Admin console. |
| `description` | `string` | — | Up to 4096 characters. |
| `security` | `bool` | `false` | Make it a security group (adds the security label; cannot be undone). |
| `initialGroupConfig` | `string` | `EMPTY` | `EMPTY` or `WITH_INITIAL_OWNER` (the caller becomes an owner). Immutable. |
| `groupNamespace` | `string` | — | `identitysources/{id}` for an identity-mapped group. Immutable. |
| `memberships` | `[]object` | — | `member` (`GcpServiceAccount` reference or email), `memberNamespace`, `roles[] { name (MEMBER / MANAGER / OWNER), expireTime (MEMBER only) }`, `createIgnoreAlreadyExists`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; applies to the memberships too. |

### Validation Rules

- Listed roles must include **`MEMBER`** and be unique; **`expireTime`** applies only to `MEMBER`.
- Each member email appears once.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `groups/{group_id}` |
| `group_email` | `string` | The group's email -- its IAM identity |
| `membership_count` | `string` | Memberships this manifest manages |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Bind IAM roles to the group, never to people.** `GcpProjectIamMember` and its siblings take `group:{groupEmail}`.
- **A security group is one-way.** The security label cannot be removed once added; make the choice at creation.
- **Deleting the group silently empties every IAM binding that named it.** `PREVENT` is the guard.
- **The group lives under the customer**, not in a project: the deploying principal needs Groups Admin, and the group outlives any project.
- **Cost**: groups and memberships are free.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the service accounts a group can hold as members
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — binds a role to `group:{groupEmail}`
- [GcpFolder](/docs/catalog/gcp/gcpfolder), [GcpProject](/docs/catalog/gcp/gcpproject) — where the group's bindings usually live

## Additional Resources

- [Cloud Identity groups](https://cloud.google.com/identity/docs/groups)
- [Create and manage security groups](https://cloud.google.com/identity/docs/how-to/update-group-to-security-group)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
