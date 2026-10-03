# Azure Redis Cache Access Policy Assignment

Grants a Redis data-plane access policy to a Microsoft Entra identity on an Azure Cache for Redis -- the Redis analog of a role assignment. The policy (built-in or an AzureRedisCacheAccessPolicy) says WHAT is allowed; this assignment says WHO gets it. This is the grant half of the keyless Redis story: a granted identity connects with its object ID (or alias) as the Redis username and an Entra token as the password -- no access key involved.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Policy Assignment** -- a grant record on the referenced cache binding one policy to one Microsoft Entra principal

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Cache for Redis** with Microsoft Entra token authentication enabled (`redisConfiguration.activeDirectoryAuthenticationEnabled: true`). Reference the AzureRedisCache Infra Component via ValueFromRef.
- **The principal being granted** -- a user, group, service principal, or managed identity. For workload identities, reference an AzureUserAssignedIdentity's `principal_id` output; for users and groups, use the literal object GUID from Entra.
- **A custom policy** (optional) -- an AzureRedisCacheAccessPolicy on the same cache, when the built-ins are too coarse.

## Deploy

### Console

Open the deployment store, find **Azure Redis Cache Access Policy Assignment**, and click **Deploy**. Start from the **Workload Identity Data Reader Grant** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRedisCacheAccessPolicyAssignment
metadata:
  name: app-identity-data-reader
  org: acme-corp
  env: prod
spec:
  redisCacheId:
    valueFrom:
      kind: AzureRedisCache
      name: app-cache
      fieldPath: status.outputs.redis_cache_id
  assignmentName: app-identity-data-reader
  accessPolicyName:
    value: "Data Reader"
  objectId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: app-identity
      fieldPath: status.outputs.principal_id
  objectIdAlias: app-identity
```

```shell
planton apply -f grant.yaml
```

This grants the built-in `Data Reader` policy to the referenced managed identity on the referenced cache -- the identity can then connect with an Entra token and read every key. An Infra Job tracks the provisioning in real time.

### InfraChart

The most natural composition is the three-kind chain -- cache, custom policy, grant -- deployed in one InfraChart:

```yaml
spec:
  redisCacheId:
    valueFrom:
      kind: AzureRedisCache
      name: app-cache
      fieldPath: status.outputs.redis_cache_id
  assignmentName: worker-session-grant
  accessPolicyName:
    valueFrom:
      kind: AzureRedisCacheAccessPolicy
      name: session-worker-policy
      fieldPath: status.outputs.access_policy_name
  objectId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: worker-identity
      fieldPath: status.outputs.principal_id
  objectIdAlias: worker-identity
```

The InfraPipeline resolves the dependency graph, provisions the cache, policy, and identity first, then creates the grant with the resolved values.

## Key Configuration

These are the most important decisions when configuring an assignment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The policy (WHAT)** -- One reference-capable field serving two modes: the three built-ins are referenced by literal name -- `Data Owner` (full access including admin commands), `Data Contributor` (read-write), `Data Reader` (read-only) -- with no policy resource existing; a custom policy is referenced through an AzureRedisCacheAccessPolicy's `access_policy_name` output.

**The principal (WHO)** -- The Entra PRINCIPAL (object) ID. For a managed identity this is the object ID, never the client ID -- granting the client ID deploys fine and fails at CONNECT time. Referencing the identity's `principal_id` output resolves the right one automatically.

**The alias** -- A human-readable label that doubles as an alternative Redis USERNAME at connect time: clients authenticate as either the raw object ID or this alias, with an Entra token as the password.

**Immutability** -- Every field is fixed at creation; any change replaces the assignment (safe: a replace momentarily revokes and re-grants).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureRedisCache** | `redisCacheId` | `status.outputs.redis_cache_id` |
| **AzureRedisCacheAccessPolicy** (optional) | `accessPolicyName` | `status.outputs.access_policy_name` |
| **AzureUserAssignedIdentity** (optional) | `objectId` | `status.outputs.principal_id` |

### What This Kind Provides

This kind has no consumable outputs: `status.outputs` carries only the assignment's ARM ID (`access_policy_assignment_id`) and its name within the cache (`access_policy_assignment_name`), and nothing downstream composes with a grant -- the assignment is a leaf in the dependency graph.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload identity, read-only** -- The everyday application grant: an AzureUserAssignedIdentity granted `Data Reader`. Start from the **Workload Identity Data Reader Grant** preset.

**Custom policy grant** -- The three-kind composition: cache, custom policy, grant. Start from the **Custom Policy Grant** preset.

**Human operator** -- A user or group GUID granted `Data Owner` for break-glass operations. Start from the **Human Operator Grant** preset.

## Works With

- [**Azure Redis Cache**](/infra-catalog/azure-redis-cache) -- the cache the grant applies to
- [**Azure Redis Cache Access Policy**](/infra-catalog/azure-redis-cache-access-policy) -- the custom WHAT half, referenced by name
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the workload principal being granted
