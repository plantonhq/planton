# AwsElasticacheUserGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsElasticacheUserGroupSpec defines the desired configuration for an AWS
ElastiCache user group — the unit of RBAC attachment for Redis/Valkey.

A user group collects AwsElasticacheUser identities and attaches, as one
object, to a replication group (`user_group_ids`) or a serverless cache
(`user_group_id`). Access control then composes through the graph: users
define WHO and WHAT (identity + ACL access string), the group defines
WHERE (which caches those identities apply to). Adding an application to a
cache is a membership edit on the group — the cache itself never changes.

AWS requires every user group to contain a user whose user name is
exactly "default": it defines what unauthenticated clients may do. The
standard production pattern is a locked-down default user (access string
"off ~* +@all", authentication "no-password-required") so anonymous
connections are rejected outright, alongside per-application users.

Notes:
- The group's AWS identifier is taken from `metadata.name` — create-time
  immutable, so renaming means replacement. `engine` is also create-time
  immutable; membership (`user_ids`) updates in place.
- A group only accepts users of its own engine and region.
- Memcached has no RBAC: user groups apply to Redis and Valkey only.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticacheUserGroup
metadata:
  name: awselasticacheusergroup-demo
spec:
  region: us-west-2
  engine: redis
  userIds:
    - value: rbac-default-user
    - value: orders-service
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engine` | `string` | yes |  |  |
| `spec.userIds` | `[]string \| valueFrom` | yes |  | AwsElasticacheUser (`status.outputs.user_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the user group is created. Must match the region
of every member user and of every cache the group attaches to.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engine

`string` · required

Cache engine this group serves. Values: "redis", "valkey". A group only
accepts users whose engine matches, and only attaches to caches of the
same engine family. Create-time immutable — changing it destroys and
recreates the group.

- rule: {"required":true}

### spec.userIds

`[]string | valueFrom` · required

The user ids that belong to this group. AWS refuses to create a group
without a user whose user NAME is "default" — include one (typically a
locked-down default user) alongside the per-application users.
Membership updates in place; the group is the single place an
application's cache access is granted or revoked.

- references: AwsElasticacheUser (`status.outputs.user_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticacheUser, name: <that resource's name>, fieldPath: status.outputs.user_id}} -- a bare string does not parse

## Validation Rules

- `engine_valid_values`: engine must be 'redis' or 'valkey' — Memcached has no RBAC user groups

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsElasticacheUserGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user_group_id` | `string` | The group's AWS identifier (the user group id). This is what caches reference to attach RBAC and what the AWS CLI/API address. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the user group. Used in IAM policies and cross-service permissions. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.userIds` | AwsElasticacheUser | `status.outputs.user_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsRedisElasticache | `spec.userGroupIds` | `status.outputs.user_group_id` |
| AwsServerlessElasticache | `spec.userGroupId` | `status.outputs.user_group_id` |

## See Also

- [Overview](../README.md)
