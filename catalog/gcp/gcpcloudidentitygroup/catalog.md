# GCP Cloud Identity Group

Creates a Google Group in Cloud Identity or Google Workspace and manages its members -- the unit every IAM binding should point at instead of individual people, so access changes are a membership edit, not a policy edit. The group is created under your Cloud Identity customer with an email in one of your domains; members are users, service accounts, or other groups by email, each with its roles and an optional expiry. Mark it a security group for access control only.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Group** -- a `cloudidentity.Group` under the customer with the declared email, display name, description, initial configuration, and the discussion-forum label every Google Group carries (plus the security label when `security` is true)
- **Memberships** -- one `cloudidentity.GroupMembership` per member, keyed by email, with its roles and optional expiry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose principal holds the Groups Admin role in the Cloud Identity / Workspace customer. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Cloud Identity

- **The customer ID** (`customers/{id}`) and a **verified domain** for the group's email.
- **Members** that already exist: Google users, groups, or `GcpServiceAccount` resources.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Identity Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Security Group for IAM** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudIdentityGroup
metadata:
  name: platform-admins
  org: acme-corp
  env: prod
spec:
  groupEmail: platform-admins@example.com
  customerId: customers/C01abc2de
  displayName: Platform Admins
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
  deletionPolicy: PREVENT
```

```shell
planton apply -f cloud-identity-group.yaml
```

This creates a security group with a human owner and a service-account member; IAM bindings then name `group:platform-admins@example.com`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire service-account members to `GcpServiceAccount` resources deployed in the same InfraPipeline, and bind roles to the group's `group_email` output through `GcpProjectIamMember`.

## Key Configuration

These are the most important decisions when configuring a group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Email and customer** -- `groupEmail` is the group's IAM identity and is immutable; pick a stable role name. `customerId` is the Cloud Identity customer.

**Security** -- `security: true` makes an access-control group (the label cannot be removed); leave it false for discussion lists.

**Memberships** -- each member by email or `GcpServiceAccount` reference with roles (`MEMBER` always; `MANAGER`, `OWNER` on top) and an optional `expireTime` on `MEMBER`. The list is the managed membership; `WITH_INITIAL_OWNER` adds the deploying principal as an owner beyond it.

**Destroy semantics** -- `deletionPolicy: PREVENT` guards a group IAM bindings depend on; deleting the group empties every binding that named it.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpServiceAccount** | `memberships[].member` | `status.outputs.email` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `groups/{group_id}` | The Cloud Identity API |
| `group_email` | The group's email | `group:{email}` in IAM bindings (`GcpProjectIamMember`) |
| `membership_count` | Memberships this manifest manages | Audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Security group for IAM** -- A security group with an owner, a manager, a service-account member, and an expiring contractor. Start from the **Security Group for IAM** preset.

**Team discussion list** -- A plain Google Group for a team's mail, with the deploying principal as the initial owner. Start from the **Team Discussion List** preset.

## Works With

- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- service accounts a group holds as members
- [**GCP Project IAM Member**](/cloud-catalog/gcp-project-iam-member) -- binds a role to `group:{group_email}`
- [**GCP Folder**](/cloud-catalog/gcp-folder), [**GCP Project**](/cloud-catalog/gcp-project) -- where the group's bindings live
