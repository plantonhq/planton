# Overview

The AwsElasticacheUserGroup API resource collects ElastiCache RBAC users
into the attachment unit for Redis and Valkey — a user group id
(`metadata.name`), an engine, and a membership list. Caches reference the
group, not individual users; granting or revoking an application's access
is a membership edit on the group.

## Why We Created This API Resource

RBAC attachment deserves its own composable node between users and caches:

- **One group, one attachment**: replication groups take `user_group_ids`;
  serverless caches take `user_group_id` — the group is the single object
  caches reference, and membership is the single place access is granted
  or revoked.
- **Compose by reference**: member users are `AwsElasticacheUser` nodes
  referenced by `status.outputs.user_id`; the group never mutates the users
  it references.
- **Stable AWS identity**: `metadata.name` is the AWS user group id
  (create-time immutable); caches reference it directly.

## Key Features

### Membership

- **`userIds`**: references to deployed users (by `user_id` output); updates
  in place — add or remove an application without touching the cache.
- **Mandatory default user**: AWS refuses group creation unless a member's
  user NAME is "default"; presets and comments steer users to include a
  locked-down default user alongside application users.

### Engine Gating

- **`engine`**: `redis` or `valkey`; must match every member user and every
  cache the group attaches to. Create-time immutable.

## Benefits

- **Composability**: users define WHO/WHAT; the group defines WHERE; caches
  reference the group — three honest nodes instead of opaque cache-side lists.
- **In-place membership**: adding an application is one more `userIds` entry;
  removing one is a list edit — no cache replacement.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `user_group_id`: the group's AWS identifier (same as `metadata.name`)
- `arn`: the group's ARN (for IAM policies and cross-service permissions)

## Deliberately Skipped (with reasons)

- **Outpost-only fields**: ElastiCache on Outposts is a separate deployment
  surface; deferred until real demand appears.
- **`aws_elasticache_user_group_association` glue resource**: membership is
  modeled as the group's `user_ids` list — the association resource has no
  independent lifecycle and is deliberately not used; updates happen in place
  on the group.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
