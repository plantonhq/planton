# AwsElasticacheUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsElasticacheUserSpec defines the desired configuration for an AWS
ElastiCache user — one identity in the Redis/Valkey RBAC (Role-Based
Access Control) system.

RBAC is AWS's recommended authentication model for ElastiCache: instead of
one shared AUTH token for every client, each application gets its own user
with an access string scoping exactly which commands and keys it may touch.
Users are grouped into user groups (AwsElasticacheUserGroup), and a user
group attaches to a replication group or serverless cache. Rotating one
application's credentials or revoking one application's access never
disturbs the others.

The user's AWS identifier (user id) is taken from `metadata.name` — it is
create-time immutable, so renaming means replacement. The `user_name` below
is the name clients present in the AUTH command; several users may share a
user name (AWS resolves the credential set across them), and every user
group must contain a user whose user name is exactly "default" — that user
defines what unauthenticated clients may do.

Memcached has no users: RBAC applies to the Redis and Valkey engines only.

Notes:
- `user_name` and `engine` are create-time immutable (ForceNew) — changing
  either destroys and recreates the user. `access_string` and the
  authentication mode update in place.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticacheUser
metadata:
  name: awselasticacheuser-demo
spec:
  region: us-west-2
  engine: redis
  userName: orders-service
  accessString: "on ~orders:* +@read +@write"
  authenticationMode:
    type: password
    passwords:
      - PlantonHackDemoPwd01!
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engine` | `string` | yes |  |  |
| `spec.userName` | `string` | yes |  |  |
| `spec.accessString` | `string` | yes |  |  |
| `spec.authenticationMode` | `AwsElasticacheUserAuthenticationMode` | yes |  |  |
| `spec.authenticationMode.type` | `string` | yes |  |  |
| `spec.authenticationMode.passwords` | `[]string` (sensitive) |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the user is created. ElastiCache users are regional
resources — a user group can only contain users from its own region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engine

`string` · required

Cache engine this user authenticates against. Values: "redis", "valkey".
A user group only accepts users whose engine matches its own.
Create-time immutable — changing it destroys and recreates the user.

- rule: {"required":true}

### spec.userName

`string` · required

The name clients present in the AUTH command (`AUTH <user_name>
<password>`). Unlike the user id (metadata.name), user names need not be
unique — AWS unions the credentials of same-named users at
authentication time. Every user group must include one user whose name
is exactly "default": it defines the permissions of clients that connect
without authenticating. Create-time immutable.

- rule: {"required":true}

### spec.accessString

`string` · required

Redis ACL access string scoping what this user may do — the same syntax
as the Redis ACL SETUSER rule list. Composed of switches ("on"/"off"
enables or disables the user), key patterns ("~app1:*" grants access to
matching keys, "~*" to all keys), and command categories ("+@read"
grants the read category, "+@all" everything, "-@dangerous" subtracts
the dangerous category).

Common shapes:
  "on ~* +@all"                      — full access (an admin user)
  "on ~app1:* +@read +@write"        — read/write scoped to one key prefix
  "on ~* +@read -@dangerous"         — read-only, no admin commands
  "off ~* +@all"                     — a disabled "default" user (clients
                                       MUST authenticate as someone else)

Updates apply in place — tightening an access string takes effect on new
connections without recreating the user.

- rule: {"required":true}

### spec.authenticationMode

`AwsElasticacheUserAuthenticationMode` · required

How clients prove they are this user. Exactly one authentication type,
with passwords carried inline only for the "password" type.

- rule: {"required":true}
- rule: authentication type must be 'password', 'iam', or 'no-password-required'
- rule: provide 1 or 2 passwords when the authentication type is 'password'
- rule: passwords are only accepted when the authentication type is 'password' — 'iam' and 'no-password-required' carry no credential material

### spec.authenticationMode.type

`string` · required

Authentication mechanism. Values:

- "password": the client presents one of the passwords below in the AUTH
  command. Two passwords may be set simultaneously so credentials rotate
  with zero downtime (add the new password, roll clients, remove the old).

- "iam": the client signs a short-lived token with its AWS IAM identity —
  no long-lived secret anywhere. AWS requires the user's `user_name` to
  equal its user id (metadata.name) for IAM auth, and the attached cache
  must have transit (TLS) encryption enabled.

- "no-password-required": the user authenticates with no credential at
  all. Intended for the mandatory "default" user when it is switched
  "off" in its access string, or for migration windows — never for a
  production user that is "on".

- rule: {"required":true}

### spec.authenticationMode.passwords

`[]string` · sensitive

Passwords for the "password" authentication type. One or two entries,
each 16–128 printable characters; two entries enable zero-downtime
rotation. Must be empty for the "iam" and "no-password-required" types.

- rule: {"repeated":{"maxItems":"2","items":{"string":{"minLen":"16","maxLen":"128"}}}}

## Validation Rules

- `engine_valid_values`: engine must be 'redis' or 'valkey' — Memcached has no RBAC users

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsElasticacheUser, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user_id` | `string` | The user's AWS identifier (the user id). This is what user groups reference in their membership list and what the AWS CLI/API address. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the user. Used in IAM policies — an IAM-authenticated client needs `elasticache:Connect` on both the user ARN and the cache ARN. |
| `status.outputs.user_name` | `string` | The name this user presents in the AUTH command. Exported so application configuration can be wired from the resource graph instead of duplicating the value. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsElasticacheUserGroup | `spec.userIds` | `status.outputs.user_id` |

## See Also

- [Overview](../README.md)
