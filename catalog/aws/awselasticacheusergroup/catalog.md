# AWS ElastiCache User Group

Deploys an ElastiCache user group — the unit of RBAC attachment for Redis and Valkey. Access control composes through a graph: users ([AwsElasticacheUser](/infra-catalog/aws-elasticache-user)) define WHO and WHAT (identity + ACL access string), and the group defines WHERE — which caches those identities apply to. The group attaches as one object to a replication group or serverless cache; granting or revoking an application's cache access is a membership edit on the group, and the cache itself never changes.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ElastiCache User Group** -- the RBAC attachment unit whose AWS user group id is the resource name (create-time immutable), carrying the set of member user ids; membership updates apply in place for the group's whole life
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Member users** -- create the [AwsElasticacheUser](/infra-catalog/aws-elasticache-user) resources first (declaration before reference); the group references their `user_id` outputs.

### AWS Account

- **A "default"-named member** -- AWS refuses to create a group unless one member's user NAME is exactly `default` (it defines what unauthenticated clients may do). The standard production shape is a locked-down default user — access string `off ~* +@all`, no credential — alongside the per-application users.
- **Matching region and engine** -- a group only accepts users of its own region and engine, and only attaches to caches that match. AWS enforces this at deploy time.

## Deploy

### Console

Open the deployment store, find **AWS ElastiCache User Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Redis User Group with Default User** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticacheUserGroup
metadata:
  name: orders-rbac
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  engine: redis
  userIds:
    - valueFrom:
        kind: AwsElasticacheUser
        name: rbac-default-user
        fieldPath: status.outputs.user_id
    - valueFrom:
        kind: AwsElasticacheUser
        name: orders-service
        fieldPath: status.outputs.user_id
```

```shell
planton apply -f elasticache-user-group.yaml
```

This creates a Redis user group wiring the mandatory default user plus one application user. An Infra Job tracks the provisioning in real time.

### InfraChart

When the group deploys alongside its member users in one chart, wire the membership via ValueFromRef:

```yaml
spec:
  region: us-west-2
  engine: redis
  userIds:
    - valueFrom:
        kind: AwsElasticacheUser
        name: rbac-default-user
        fieldPath: status.outputs.user_id
    - valueFrom:
        kind: AwsElasticacheUser
        name: orders-service
        fieldPath: status.outputs.user_id
```

The InfraPipeline resolves the dependency graph, deploys the users first, then this group -- and the cache that attaches it references the `user_group_id` output afterwards.

## Key Configuration

These are the most important decisions when configuring an ElastiCache user group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The resource name IS the AWS user group id** -- what caches reference in their user-group fields, and create-time immutable (renaming means replacement).

**Engine is a one-way door** -- a group serves exactly one engine family (`redis` or `valkey`; Memcached has no RBAC). Changing it destroys and recreates the group, detaching it from every cache until redeployed. Members and attached caches must match it.

**Membership is the operational lever** -- `userIds` updates in place: adding an application to a cache is adding its user id here, revoking is removing it. Prefer member entries by reference (the user's `user_id` output) so a renamed or recreated user fails loudly at deploy instead of silently pointing at nothing; literal ids stay first-class for users created outside the platform.

**Ids, not names** -- membership binds user IDs (unique), never AUTH user names (which several users may share). AWS also rejects a group listing the same user id twice.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsElasticacheUser** | `userIds` | `status.outputs.user_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `user_group_id` | The group's AWS identifier (the resource name) | Serverless cache `userGroupId`, replication group `user_group_ids` |
| `arn` | Amazon Resource Name of the user group | IAM policies and cross-service permissions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Redis RBAC group** -- the locked-down default user plus one user per application service, attached to a Redis replication group or serverless cache. Start from the **Redis User Group with Default User** preset.

**Valkey RBAC group** -- the same membership model for Valkey caches; every member must be a valkey-engine user. Start from the **Valkey User Group** preset.

**Per-environment groups** -- one group per environment (dev/staging/prod) with different membership, so a staging credential can never reach the production cache.

## Works With

- [**AWS ElastiCache User**](/infra-catalog/aws-elasticache-user) -- the identities this group collects (consumes `user_id`)
- [**AWS ElastiCache Serverless**](/infra-catalog/aws-serverless-elasticache) -- attaches this group via its user group field (`user_group_id`)
