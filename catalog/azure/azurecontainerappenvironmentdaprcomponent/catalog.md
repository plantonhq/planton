# Azure Container App Environment Dapr Component

Registers a Dapr component on a Container App Environment -- the pluggable backend behind one of Dapr's building blocks: a state store (state.azure.blobstorage, state.redis), a pub/sub broker (pubsub.azure.servicebus), a secret store, or a binding. Application code calls the Dapr API with the component's NAME; which database or broker actually serves the call is this registration's decision, swappable without touching application code. Components register once on the environment and are consumed by any Dapr-enabled app whose `dapr.app_id` appears in `scopes` -- an empty scopes list exposes the component to every Dapr-enabled app in the environment.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Dapr Component** -- on the referenced Container App Environment, with its type and version, the metadata configuration entries, the component-scoped secrets those entries may reference, and the app-ID scopes

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureContainerAppEnvironment** to register on. Reference its `environment_id` output via ValueFromRef.
- **The backing service** the component type addresses -- a storage account for state.azure.blobstorage, a Service Bus namespace for pubsub.azure.servicebus, a Key Vault for secretstores.azure.keyvault.
- **For keyless auth** -- an AzureUserAssignedIdentity with the data-plane role on the backend, referenced from an `azureClientId` metadata entry (the alternative to connection-string secrets).

## Deploy

### Console

Open the deployment store, find **Azure Container App Environment Dapr Component**, and click **Deploy**. The creation wizard walks you through the component identity (environment, the name application code calls, and the type as a curated free-solo over Dapr's dotted notation), the runtime dials (version, init timeout, the fail-loudly error posture), the declare-before-reference secrets, the per-entry value-or-secret configuration, and the app-ID scopes with the empty-means-everyone exposure warning. Start from the **Blob Storage State Store** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppEnvironmentDaprComponent
metadata:
  name: orders-state-component
  org: acme-corp
  env: prod
spec:
  containerAppEnvironmentId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.App/managedEnvironments/prod-apps-env"
  componentName: orders-state
  componentType: state.azure.blobstorage
  version: v1
  secrets:
    - name: account-key
      value: $secret/prod-storage-key
  metadata:
    - name: accountName
      value:
        value: proddata001
    - name: containerName
      value:
        value: orders-state
    - name: accountKey
      secretName: account-key
  scopes:
    - orders-api
```

```shell
planton apply -f component.yaml
```

This registers a blob-storage state store named `orders-state`, scoped to the `orders-api` app, with the account key delivered as a managed secret. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the component to the environment -- and inside metadata entries to wire deploy-time values, where the keyless-auth pattern is the canonical case:

```yaml
spec:
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: apps-env
      fieldPath: status.outputs.environment_id
  metadata:
    - name: azureClientId
      value:
        valueFrom:
          kind: AzureUserAssignedIdentity
          name: orders-identity
          fieldPath: status.outputs.client_id
```

The InfraPipeline resolves the dependency graph, deploys the environment, identity, and backend first, then registers the component.

## Key Configuration

These are the most important decisions when configuring a component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Name and type** -- the name is what code calls (name it for the ROLE: orders-state, not orders-blob-storage); the type names the building block AND the backend in Dapr's dotted notation. Both are fixed at creation, along with the environment -- swapping the backend later is a new component plus a data migration, and Dapr's promise is that the migration never touches application code. Everything else -- version, timeout, error posture, secrets, metadata, scopes -- edits in place; sidecars pick up the change when they restart. The type deliberately carries no pinned vocabulary: any component the Dapr runtime ships stays expressible.

**Metadata** -- the component's configuration, with keys defined by the chosen type's Dapr reference. Each entry is a VALUE (a literal, or a reference for deploy-time facts like a client ID) or a SECRET reference -- never both; connection strings and keys always travel through the secrets list.

**Scopes** -- the exposure boundary. Empty exposes the component (and whatever its secrets unlock) to EVERY Dapr-enabled app in the environment; entries are `dapr.app_id` values, not Container App resource names. Scope production components deliberately; scopes edit in place.

**Failure posture** -- `ignoreErrors` defaults to false, the safe choice: a broken component stops the sidecar loudly at startup instead of surfacing as runtime errors on first use.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |
| **AzureUserAssignedIdentity** | `metadata[].value` (keyless auth) | `status.outputs.client_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `component_name` | The component's name on the environment | What application code passes to the Dapr API -- there is no endpoint; the sidecar resolves it |
| `dapr_component_id` | Azure Resource Manager ID of the component | Operational tooling and portal navigation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Blob state store** -- state.azure.blobstorage with a secret-backed account key, scoped to the owning app. Start from the **Blob Storage State Store** preset.

**Shared pub/sub** -- pubsub.azure.servicebus registered unscoped as environment-wide messaging infrastructure, authenticated keylessly through an `azureClientId` entry. Start from the **Service Bus Pub/Sub (Keyless)** preset.

**Promotion-path parity** -- staging and production environments each carry their own registration of the same logical component: same name, same type, different backing account -- application code stays identical across the promotion path.

## Works With

- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- where the component registers
- [**Azure Container App**](/infra-catalog/azure-container-app) -- Dapr-enabled apps consume the component by name, scoped via their `dapr.app_id`
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the classic state-store backend
- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- the classic pub/sub backend
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the secret-store backend
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- keyless component authentication via the azureClientId metadata entry
