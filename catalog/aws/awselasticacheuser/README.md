# Overview

The AwsElasticacheUser API resource creates one identity in the ElastiCache
RBAC (Role-Based Access Control) system for Redis and Valkey — a named user
with an ACL access string and an authentication mode. Users compose into
user groups, and groups attach to caches; rotating one application's
credentials never disturbs the others.

## Why We Created This API Resource

Per-application cache credentials deserve a first-class, composable node:

- **One user, one node**: each application gets its own identity with an
  access string scoping exactly which keys and commands it may touch; the
  architecture graph shows WHO may reach the cache and WHAT they may do.
- **Honest auth modes**: password (with optional dual-password rotation),
  IAM-signed tokens, or no-password-required (for the mandatory locked-down
  "default" user) — CEL enforces exactly one mode with the right credential
  material.
- **Stable AWS identity**: `metadata.name` is the AWS user id (create-time
  immutable); user groups reference it in their membership lists.

## Key Features

### Authentication Modes

- **Password**: clients present one of 1–2 passwords in the AUTH command;
  two passwords enable zero-downtime rotation.
- **IAM**: clients sign short-lived tokens with their AWS IAM identity — no
  long-lived secret; requires `userName` to equal the user id and transit
  encryption on the attached cache.
- **No-password-required**: for the mandatory "default" user when its access
  string is switched "off" — rejects unauthenticated clients outright.

### Access Control

- **Redis ACL access strings**: key patterns (`~app:*`) and command categories
  (`+@read`, `-@dangerous`, `+@all`) — the same syntax as Redis ACL SETUSER.
- **In-place updates**: `accessString` and authentication mode change without
  recreating the user; `userName` and `engine` are create-time immutable.

## Benefits

- **Composability**: user groups reference users by `status.outputs.user_id`;
  caches reference groups — grant or revoke access by editing membership,
  never by touching the cache.
- **Honest constraints**: engine gating (redis/valkey only), auth-type
  coupling, and password count limits are CEL-enforced at validation time.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `user_id`: the user's AWS identifier (same as `metadata.name`)
- `arn`: the user's ARN (for IAM `elasticache:Connect` policies)
- `user_name`: the name clients present in the AUTH command

## Deliberately Skipped (with reasons)

- **Outpost-only fields** (`user_id` overrides, Outpost ARNs): ElastiCache
  on Outposts is a separate deployment surface with no current Planton demand;
  deferred until a concrete use case appears.
- **Legacy flat auth arms** (`passwords` / `no_password_required` top-level
  fields on the provider resource): the nested `authentication_mode` block is
  the single honest shape — one type discriminator with credential material
  only on the "password" arm.
- **Terraform write-only `passwords_wo`**: the spec carries passwords inline
  (marked sensitive) for both engines; write-only password fields are a
  Terraform-provider convenience with no protobuf equivalent and are not
  modeled.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
