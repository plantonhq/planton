# AwsMemorydbUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsMemorydbUserSpec defines the desired configuration for an AWS MemoryDB
user — one identity in MemoryDB's Access Control List (ACL) authentication
system.

MemoryDB has exactly one authentication model: every cluster attaches an
ACL, and an ACL is a set of users. There is no AUTH-token mode and no
security-group-only mode — if an application should reach a MemoryDB
cluster with credentials, a user is how that identity exists. Each user
carries a Redis ACL access string scoping which commands and keys it may
touch, so per-application least-privilege access is the natural shape:
one user per application, grouped into ACLs (AwsMemorydbAcl), with the
ACL attached to the cluster.

The user's name — the identity clients present in the AUTH command and the
identifier ACLs reference — is taken from `metadata.name`. Unlike
ElastiCache (where a user has a separate id and a reusable user name),
MemoryDB has a single flat user name that IS the identity: it must be
unique in the region and is create-time immutable, so renaming means
replacement.

Notes:
- AWS caps the user name at 40 characters; longer `metadata.name` values
  are rejected by AWS at create time.
- `access_string` and the authentication mode update in place — tightening
  permissions or rotating passwords never recreates the user.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemorydbUser
metadata:
  name: orders-service
  id: test-memorydb-user
  org: test-org
  env: dev
spec:
  region: us-west-2
  accessString: on ~orders:* +@read +@write
  authenticationMode:
    type: password
    passwords:
      - a-very-strong-password
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.accessString` | `string` | yes |  |  |
| `spec.authenticationMode` | `AwsMemorydbUserAuthenticationMode` | yes |  |  |
| `spec.authenticationMode.type` | `string` | yes |  |  |
| `spec.authenticationMode.passwords` | `[]string` (sensitive) |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the user is created. MemoryDB users are regional
resources — an ACL can only contain users from its own region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.accessString

`string` · required

Redis ACL access string scoping what this user may do — the same syntax
as the Redis ACL SETUSER rule list. Composed of switches ("on"/"off"
enables or disables the user), key patterns ("~app1:*" grants access to
matching keys, "~*" to all keys), and command categories ("+@read"
grants the read category, "+@all" everything, "-@dangerous" subtracts
the dangerous category).

Common shapes:
  "on ~* +@all"                — full access (an admin user)
  "on ~app1:* +@read +@write"  — read/write scoped to one key prefix
  "on ~* +@read -@dangerous"   — read-only, no admin commands

Updates apply in place — a tightened access string takes effect on new
connections without recreating the user.

AWS stores the string NORMALIZED (live-verified 2026-08-13): the API
echoes it with Redis ACL defaults made explicit — e.g.
"on ~orders:* +@read" comes back as
"on ~orders:* resetchannels -@all +@read". The provider reconciles the
normalized form, so manifests never see drift; just don't expect a
byte-identical echo when comparing DescribeUsers output to this field.

- rule: {"required":true}

### spec.authenticationMode

`AwsMemorydbUserAuthenticationMode` · required

How clients prove they are this user. Exactly one authentication type,
with passwords carried inline only for the "password" type.

- rule: {"required":true}
- rule: authentication type must be 'password' or 'iam' — MemoryDB users have no passwordless type
- rule: provide 1 or 2 passwords when the authentication type is 'password'
- rule: passwords are only accepted when the authentication type is 'password' — 'iam' authenticates with short-lived signed tokens instead

### spec.authenticationMode.type

`string` · required

Authentication mechanism. Values:

- "password": the client presents one of the passwords below in the AUTH
  command. Two passwords may be set simultaneously so credentials rotate
  with zero downtime (add the new password, roll clients, remove the old).

- "iam": the client signs a short-lived token with its AWS IAM identity —
  no long-lived secret anywhere. AWS requires TLS on the cluster for
  IAM-authenticated connections, and the IAM principal needs
  `memorydb:Connect` on both the user ARN and the cluster ARN.

- rule: {"required":true}

### spec.authenticationMode.passwords

`[]string` · sensitive

Passwords for the "password" authentication type. One or two entries,
each 16–128 characters; two entries enable zero-downtime rotation.
Must be empty for the "iam" type. Write-only at AWS: the API never
returns passwords, so drift changed outside this manifest is
undetectable, and a user adopted by import carries no password state —
re-assert the passwords here to manage them.

- rule: {"repeated":{"maxItems":"2","items":{"string":{"minLen":"16","maxLen":"128"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsMemorydbUser, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user_name` | `string` | The user's name — the single identity clients present in the AUTH command and the identifier ACLs reference in their membership list. Exported so ACL membership and application configuration are wired from the resource graph instead of duplicating the value. |
| `status.outputs.user_arn` | `string` | The Amazon Resource Name of the user. Used in IAM policies — an IAM-authenticated client needs `memorydb:Connect` on both the user ARN and the cluster ARN. |
| `status.outputs.minimum_engine_version` | `string` | The minimum engine version the user's configuration requires. AWS computes this from the ACL feature set the user exercises; an ACL (and the cluster it attaches to) must run at least this engine version. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsMemorydbAcl | `spec.userNames` | `status.outputs.user_name` |

## See Also

- [Overview](../README.md)
