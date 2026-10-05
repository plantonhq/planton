# Azure Redis Cache Access Policy

Defines a CUSTOM data-plane access policy on an Azure Cache for Redis -- a named permission set, written in Redis's own ACL syntax, that Microsoft Entra identities are granted through AzureRedisCacheAccessPolicyAssignment. This is the Redis data-plane analog of a custom role definition: the policy says WHAT is allowed (commands, command categories, key patterns); the assignment says WHO gets it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Policy** -- a named permission set on the referenced cache, expressed in Redis ACL syntax and updatable in place

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Cache for Redis** with Microsoft Entra token authentication enabled (`redisConfiguration.activeDirectoryAuthenticationEnabled: true`) -- policies gate token-authenticated clients, not access-key clients. Reference the AzureRedisCache Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Redis Cache Access Policy**, and click **Deploy**. Start from the **Read-Only Policy on a Key Prefix** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRedisCacheAccessPolicy
metadata:
  name: session-worker-policy
  org: acme-corp
  env: prod
spec:
  redisCacheId:
    valueFrom:
      kind: AzureRedisCache
      name: app-cache
      fieldPath: status.outputs.redis_cache_id
  policyName: session-worker
  permissions: "+get +set +del ~session:*"
```

```shell
planton apply -f access-policy.yaml
```

This creates a custom policy named `session-worker` on the referenced cache, allowing exactly three commands (`GET`, `SET`, `DEL`) on keys under the `session:` prefix. An Infra Job tracks the provisioning in real time.

### InfraChart

When the policy and its cache deploy in the same InfraChart, wire the cache reference with ValueFromRef:

```yaml
spec:
  redisCacheId:
    valueFrom:
      kind: AzureRedisCache
      name: app-cache
      fieldPath: status.outputs.redis_cache_id
  policyName: session-worker
  permissions: "+get +set +del ~session:*"
```

The InfraPipeline resolves the dependency graph, provisions the cache first, then creates the policy on it with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring an access policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Built-in policy or custom policy** -- Azure ships three built-in policies -- **Data Owner** (full access including admin commands), **Data Contributor** (read-write), and **Data Reader** (read-only) -- which AzureRedisCacheAccessPolicyAssignment references by literal name without any policy resource existing. Create a custom policy only when the built-ins are too coarse: read-write on one key prefix, no admin commands, a three-command worker. If a built-in fits, skip this kind entirely and write the assignment directly.

**Permissions (Redis ACL syntax)** -- Command/category grants followed by key patterns. Building blocks: `+@read` allows a command category, `+get` a single command, `-@dangerous` carves one out, and `~pattern` scopes which keys the grants apply to (`~*` for all keys). Semantics are ALLOW-only with carve-outs -- there is no deny-first mode. Updatable in place: an edit changes what every assignment holding this policy can do, immediately.

**Policy name** -- What assignments reference, unique within the cache. Must not collide with the built-in names. Changing the name replaces the policy and strands assignments holding the old one.

**Cache binding** -- `redisCacheId` is fixed at creation: a policy cannot move between caches. Reorganizing caches means recreating every custom policy (and its assignments) on the new cache -- keep the policy manifests alongside the cache manifest so the whole set replays together.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureRedisCache** | `redisCacheId` | `status.outputs.redis_cache_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `access_policy_name` | The policy's name | AzureRedisCacheAccessPolicyAssignment's `accessPolicyName` |

The outputs also carry `access_policy_id`, the policy's ARM ID -- assignments reference the policy by name within the cache, so the ID has no ValueFromRef consumer.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Read-only on a prefix** -- `+@read +@connection ~metrics:*` for dashboards and debuggers. Start from the **Read-Only Policy on a Key Prefix** preset.

**Application writer, no admin** -- `+@all -@dangerous ~app1:*` for the everyday application identity. Start from the **Application Writer Without Admin Commands** preset.

**Least-privilege worker** -- exact commands on one prefix, e.g. `+get +set +del ~session:*`. Start from the **Queue Worker with Single-Command Grants** preset.

## Works With

- [**Azure Redis Cache**](/infra-catalog/azure-redis-cache) -- the cache the policy is defined on
- [**Azure Redis Cache Access Policy Assignment**](/infra-catalog/azure-redis-cache-access-policy-assignment) -- grants this policy (by name) to a Microsoft Entra identity
