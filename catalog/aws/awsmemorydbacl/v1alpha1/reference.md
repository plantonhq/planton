# AwsMemorydbAcl

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsMemorydbAclSpec defines the desired configuration for an AWS MemoryDB
Access Control List (ACL) — the set of users a MemoryDB cluster
authenticates against.

An ACL is the single attachment point between identities and clusters:
users (AwsMemorydbUser) are grouped into an ACL, and a cluster attaches
exactly one ACL. Granting or revoking an application's access to a cluster
is an ACL membership edit — in place, with no cluster or user replacement.
One ACL is typically shared by every cluster in an environment that trusts
the same set of application identities.

The ACL's AWS name is taken from `metadata.name` — create-time immutable,
so renaming means replacement. Membership (`user_names`) updates in place.

Notes:
- AWS caps the ACL name at 40 characters; longer `metadata.name` values
  are rejected by AWS at create time.
- An empty ACL is valid (unlike ElastiCache user groups, MemoryDB has no
  mandatory "default" member) — but a cluster attached to an empty ACL
  accepts no authenticated connections, so production ACLs carry one user
  per application.
- The built-in "open-access" ACL (every command, every key, no
  authentication) always exists in the account and is referenced by name —
  it is never modeled as a resource.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemorydbAcl
metadata:
  name: payments-env-acl
  id: test-memorydb-acl
  org: test-org
  env: dev
spec:
  region: us-west-2
  userNames:
    - value: orders-service
    - value: analytics-service
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.userNames` | `[]string \| valueFrom` |  |  | AwsMemorydbUser (`status.outputs.user_name`) |

## Field Details

### spec.region

`string` · required

The AWS region where the ACL is created. Must match the region of every
member user and of every cluster the ACL attaches to.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.userNames

`[]string | valueFrom`

The users that make up this ACL, referenced by user name (the value
AwsMemorydbUser exports as `status.outputs.user_name`). Membership
updates in place; the ACL is the single place an application's cluster
access is granted or revoked. May be empty — a cluster attached to an
empty ACL simply accepts no authenticated connections. Deleting a user
removes it from every ACL server-side automatically; the provider
reconciles that, so dropping an already-deleted member from this list
is a clean no-op, never an error.

- references: AwsMemorydbUser (`status.outputs.user_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsMemorydbUser, name: <that resource's name>, fieldPath: status.outputs.user_name}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsMemorydbAcl, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.acl_name` | `string` | The ACL's AWS name. This is what clusters attach via their `acl_name` and what the AWS CLI/API address. |
| `status.outputs.acl_arn` | `string` | The Amazon Resource Name of the ACL. Used in IAM policies and cross-service permissions. |
| `status.outputs.minimum_engine_version` | `string` | The minimum engine version the ACL's combined user set requires. A cluster attaching this ACL must run at least this engine version. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.userNames` | AwsMemorydbUser | `status.outputs.user_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsMemorydbCluster | `spec.aclName` | `status.outputs.acl_name` |

## See Also

- [Overview](../README.md)
