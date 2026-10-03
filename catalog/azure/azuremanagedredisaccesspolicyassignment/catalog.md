# Azure Managed Redis Access Policy Assignment

Grants Redis data-plane access to a Microsoft Entra identity on an Azure Managed Redis instance -- the Redis analog of a role assignment, and the grant half of Managed Redis's keyless-by-default story. Access keys are off unless explicitly enabled, so a granted identity is how clients connect at all: the identity presents its object ID as the Redis username and an Entra token as the password, and no secret exists to leak or rotate.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Access Policy Assignment** -- a grant on the instance's default database, binding the Entra principal to the built-in "default" access policy (full data access). Azure names the assignment after the granted object ID, so an identity is granted at most once per database.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Managed Redis instance** to grant on. Reference an AzureManagedRedis Infra Component via ValueFromRef, or provide the ARM ID directly.
- **The Entra principal's object ID** -- a managed identity, service principal, user, or group. For a managed identity this is the PRINCIPAL (object) id, never the client id; granting the client id fails at connect time, not at deploy time.

## Deploy

### Console

Open the deployment store, find **Azure Managed Redis Access Policy Assignment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the grant itself. Start from the **Workload Identity Grant** preset in the [Presets](#presets) tab for the secretless workload pattern.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureManagedRedisAccessPolicyAssignment
metadata:
  name: app-cache-grant
  org: acme-corp
  env: prod
spec:
  managedRedisId:
    valueFrom:
      kind: AzureManagedRedis
      name: app-cache
      fieldPath: status.outputs.managed_redis_id
  objectId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: app-identity
      fieldPath: status.outputs.principal_id
```

```shell
planton apply -f grant.yaml
```

This grants the identity full data-plane access on the instance's default database; the identity connects with its object ID as the username and an Entra token as the password. An Infra Job tracks the provisioning in real time.

### InfraChart

The grant is a natural InfraChart edge -- both fields reference sibling resources, so the pipeline deploys the instance and identity first, then the grant:

```yaml
spec:
  managedRedisId:
    valueFrom:
      kind: AzureManagedRedis
      name: app-cache
      fieldPath: status.outputs.managed_redis_id
  objectId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: app-identity
      fieldPath: status.outputs.principal_id
```

## Key Configuration

These are the only two decisions -- the kind is deliberately minimal. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The instance** -- `managedRedisId` by ARM ID; the grant is created on its default database.

**The principal** -- `objectId` is the GUID being granted. Workload identities reference an AzureUserAssignedIdentity's `principal_id` output; humans and groups pass a literal GUID from Microsoft Entra ID. Managed Redis exposes one built-in policy ("default", full data access) -- every assignment grants it, and there is nothing else to name or configure. Every field is fixed at creation: changing anything replaces the assignment (safe -- a replace momentarily revokes and re-grants).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureManagedRedis** | `managedRedisId` | `status.outputs.managed_redis_id` |
| **AzureUserAssignedIdentity** (workload grants) | `objectId` | `status.outputs.principal_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `access_policy_assignment_id` | Azure Resource Manager ID of the assignment | Audit and cross-referencing |
| `access_policy_assignment_name` | The assignment's name -- equals the granted object ID | The Redis USERNAME the identity connects with |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload identity grant** -- The secretless application pattern: a user-assigned identity granted per instance, referenced end to end. Start from the **Workload Identity Grant** preset.

**Human operator grant** -- A user or Entra group (covering a whole on-call rotation with one assignment) granted by literal GUID -- personal, auditable access with no shared key in a vault. Start from the **Human Operator Grant** preset.

## Works With

- [**Azure Managed Redis**](/infra-catalog/azure-managed-redis) -- the instance being granted on
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the workload identity whose principal_id is granted
- [**Azure Managed Redis Geo Replication**](/infra-catalog/azure-managed-redis-geo-replication) -- grants are per instance; geo-replicated applications grant their identity on every member
