# Overview

The AwsMemorydbUser API resource creates one identity in MemoryDB's ACL
authentication system — a named user with a Redis ACL access string and an
authentication mode. Users compose into ACLs (AwsMemorydbAcl), and a cluster
attaches exactly one ACL; rotating one application's credentials never
disturbs the others.

## Why We Created This API Resource

MemoryDB has exactly one authentication model — ACLs made of users — so
per-application credentials deserve a first-class, composable node:

- **One user, one node**: each application gets its own identity with an
  access string scoping exactly which keys and commands it may touch; the
  architecture graph shows WHO may reach the database and WHAT they may do.
- **Honest auth modes**: password (with optional dual-password rotation) or
  IAM-signed tokens — CEL enforces exactly one mode with the right
  credential material. MemoryDB has no passwordless user type;
  unauthenticated access exists only through the built-in "open-access" ACL.
- **Stable AWS identity**: `metadata.name` IS the AWS user name (the AUTH
  identity, create-time immutable); ACLs reference it in their membership
  lists via `status.outputs.user_name`.

## Key Features

### Authentication Modes

- **Password**: clients present one of 1–2 passwords (16–128 characters) in
  the AUTH command; two passwords enable zero-downtime rotation (add the new
  password, roll clients, remove the old). Write-only at AWS: the API never
  returns passwords, so out-of-band changes are undetectable and imported
  users carry no password state — the manifest is the source of truth.
- **IAM**: clients sign short-lived tokens with their AWS IAM identity — no
  long-lived secret anywhere; requires TLS on the cluster and
  `memorydb:Connect` on both the user ARN and the cluster ARN.

### Access Control

- **Redis ACL access strings**: key patterns (`~app:*`) and command
  categories (`+@read`, `-@dangerous`, `+@all`) — the same syntax as Redis
  ACL SETUSER.
- **In-place updates**: `accessString` and the authentication mode change
  without recreating the user; the user name (`metadata.name`) is
  create-time immutable.

## Benefits

- **Composability**: ACLs reference users by `status.outputs.user_name`;
  clusters reference ACLs — grant or revoke access by editing ACL
  membership, never by touching the cluster.
- **Secret-safe**: passwords are marked sensitive and never appear in
  rendered manifests or logs.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `user_name`: the AUTH identity ACLs reference (same as `metadata.name`)
- `user_arn`: the user's ARN (for IAM `memorydb:Connect` policies)
- `minimum_engine_version`: the minimum engine version the user's
  configuration requires

## Deliberately Skipped (with reasons)

- **`no-password-required` user type**: unlike ElastiCache, MemoryDB's
  CreateUser input accepts only `password` and `iam` — a passwordless user
  does not exist in the API, so the spec does not invent one.
- **Separate `user_name` spec field**: in MemoryDB the user name IS the
  user's single identity (there is no ElastiCache-style user-id/user-name
  split), so it derives from `metadata.name` per the naming convention.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
