# Azure Mongo Cluster User

Grants a Microsoft Entra ID principal access to an Azure Cosmos DB for MongoDB vCore cluster -- an access binding, not a password user. The principal (a person, a service principal, or a managed identity's service principal) authenticates to MongoDB with its Entra token over the `MONGODB-OIDC` mechanism and receives the granted database roles: no password to store, rotate, or leak, and revoking access is deleting this resource. Every field is create-only -- changing anything replaces the grant, a harmless drop-and-re-add of an access binding.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Mongo cluster user grant** -- the Entra principal's binding on the cluster, with its database role grants. The identity provider is pinned to `MicrosoftEntraID` (the only value Azure accepts today), so it never appears in the spec -- both engines send it explicitly. The grant carries no tags: ARM data-plane user entries are untagged.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **The target cluster** -- reference an AzureMongoCluster's `mongo_cluster_id` output, or pass an existing cluster's ARM ID.
- **The principal** -- for workload identities, reference the AzureUserAssignedIdentity's `principal_id` output; for humans or app registrations, the object ID from Entra.

### Azure Subscription

- **The cluster must allow Entra authentication** -- "MicrosoftEntraID" must be listed in the cluster's `authenticationMethods`; Azure rejects the grant at deploy time otherwise. The Free tier refuses Entra authentication entirely, so the cheapest cluster a grant can target is M10.
- **A grant is free** -- it is a data-plane user entry, not billable infrastructure.

## Deploy

### Console

Open the deployment store, find **Azure Mongo Cluster User**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the target cluster, the principal and its type, and the roles. Start from the **App Identity Access** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMongoClusterUser
metadata:
  name: orders-app-access
  org: acme-corp
  env: prod
spec:
  mongoClusterId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-data/providers/Microsoft.DocumentDB/mongoClusters/acme-orders-db
  objectId:
    value: 11111111-2222-3333-4444-555555555555
  principalType: servicePrincipal
  roles:
    - database: admin
      role: root
```

```shell
planton apply -f mongo-cluster-user.yaml
```

This grants the service principal behind the object ID cluster-wide access -- root on the admin database -- to the named Mongo vCore cluster. An Infra Job tracks the provisioning in real time.

### InfraChart

When the grant's dependencies deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  mongoClusterId:
    valueFrom:
      kind: AzureMongoCluster
      name: orders-db
      fieldPath: status.outputs.mongo_cluster_id
  objectId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: orders-app-identity
      fieldPath: status.outputs.principal_id
  principalType: servicePrincipal
  roles:
    - database: admin
      role: root
```

The InfraPipeline resolves the dependency graph, deploys the cluster and the identity first, then lands the grant -- the whole passwordless wiring in one chart.

## Key Configuration

These are the most important decisions when configuring a Mongo cluster user grant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Grants are how applications connect; the administrator is break-glass** -- the cluster's native administrator credential is one shared password with total power, and an application that ships with it turns every password rotation into an every-consumer outage. Wire each application through its own grant instead: the identity authenticates with short-lived Entra tokens, and each binding revokes alone by deleting its resource.

**Get the RIGHT UUID into `objectId`** -- managed identities carry two UUIDs, the client ID and the PRINCIPAL (object) ID, and the grant wants the principal ID: reference the identity's `principal_id` output rather than copying UUIDs by hand. For app registrations, use the ENTERPRISE APPLICATION's object ID (the service principal in your tenant), not the app registration's object ID -- the wrong one creates a grant nobody can authenticate against, which looks like a driver bug, not a config bug.

**`principalType` names what the object ID is** -- "servicePrincipal" for applications and managed identities (identities bind through their service principal), "user" for people. It is create-only like everything else here, so a mismatch is fixed by a replace, not an update.

**Replacement is the update model, and that is fine** -- every field is create-only: role changes, principal changes, anything drops and re-adds the grant. That is safe for an access binding (no data rides on it), but there is a moment without access during the replace -- for zero-gap changes on a live application, add a second grant first, then remove the old one.

**Entra auth is a cluster-level switch someone must have thrown** -- a grant against a cluster whose `authenticationMethods` lacks "MicrosoftEntraID" fails at deploy time. When adopting Entra auth on an existing NativeAuth-only cluster, update the cluster first (an in-place update), then land the grants -- and keep "NativeAuth" in the list during the migration, because removing it cuts over every consumer at once.

**The role vocabulary is "root" today; the real decision is the database** -- Azure's Mongo vCore service accepts exactly one role name (the provider rejects anything else), so scope is chosen by the `database` the role is granted on: "admin" is cluster-wide access. Azure owns the vocabulary and will widen it over time.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureMongoCluster** | `mongoClusterId` | `status.outputs.mongo_cluster_id` |
| **AzureUserAssignedIdentity** (workload identities) | `objectId` | `status.outputs.principal_id` |

### What This Kind Provides

This kind has no consumable outputs. `status.outputs` records the grant's ARM ID (`mongo_cluster_user_id`, shaped `{cluster_id}/users/{object_id}`) and its ARM name (`mongo_cluster_user_name`, which is the granted principal's object ID back), but both derive from the inputs -- applications connect to the CLUSTER's endpoints under the granted identity, not to the grant, so nothing downstream references these values.

## Common Patterns

**Passwordless application access** -- grant an application's managed identity root on its cluster and let it connect with `MONGODB-OIDC` under its Entra token: nothing to rotate, and each app's access is revocable on its own. Start from the **App Identity Access** preset.

**Break-glass administrator posture** -- wire every application through a grant and keep the cluster's native administrator credential vaulted for incidents; the grant model exists precisely so the shared password never lands in application configuration.

**Zero-gap role changes** -- because every change replaces the grant, a role migration on a live application lands as two applies: add a second grant carrying the new roles, confirm the application still authenticates, then delete the old grant.

## Works With

- [**Azure Mongo Cluster (Cosmos DB for MongoDB vCore)**](/infra-catalog/azure-mongo-cluster) -- the cluster the principal is granted access to; it must list "MicrosoftEntraID" in its authentication methods
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the workload identity whose principal the grant binds, referenced by its `principal_id` output
