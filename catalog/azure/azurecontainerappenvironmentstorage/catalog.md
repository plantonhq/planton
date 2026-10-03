# Azure Container App Environment Storage

Registers an Azure Files share as a named storage resource on a Container App Environment -- the mount bridge of the Container Apps family. Apps and jobs cannot mount a file share directly: the share is first registered on the environment, and workload volumes then declare AZURE_FILE (SMB) or NFS_AZURE_FILE volumes that reference the registration by its `storageName`. One registration can back volumes in any number of apps and jobs in the environment, over exactly one protocol: SMB (the common case) or NFS (VNet-injected environments only).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Environment Storage Registration** -- on the referenced Container App Environment, addressing the share over exactly one protocol: SMB (storage account name + access key) or NFS (the account's file endpoint, for VNet-injected environments)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureContainerAppEnvironment** to register on. Reference its `environment_id` output via ValueFromRef.
- **An AzureStorageShare** (and its AzureStorageAccount) holding the data. For NFS, the share must be an NFS-protocol share on a premium FileStorage account, and the environment must be VNet-injected -- NFS traffic never leaves the virtual network.

## Deploy

### Console

Open the deployment store, find **Azure Container App Environment Storage**, and click **Deploy**. The creation wizard walks you through the registration identity (environment + the name workload volumes reference), the share and its protocol (with a same-account quick-fill wiring the SMB key from the chosen account), and the read-write-or-read-only access mode. Start from the **SMB Azure Files Share (Read-Write)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerAppEnvironmentStorage
metadata:
  name: app-data-registration
  org: acme-corp
  env: prod
spec:
  containerAppEnvironmentId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.App/managedEnvironments/prod-apps-env"
  storageName: app-data
  shareName:
    value: "app-data-share"
  accessMode: READ_WRITE
  accountName:
    value: "acmeproddata001"
  accessKey:
    value: $secret/prod-storage-primary-key
```

```shell
planton apply -f storage.yaml
```

This registers the SMB share `app-data-share` on the environment as `app-data`, mountable read-write by any app or job volume that references it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the registration to the environment, share, and account deployed in the same InfraPipeline:

```yaml
spec:
  containerAppEnvironmentId:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: apps-env
      fieldPath: status.outputs.environment_id
  storageName: app-data
  shareName:
    valueFrom:
      kind: AzureStorageShare
      name: app-data-share
      fieldPath: status.outputs.share_name
  accessMode: READ_WRITE
  accountName:
    valueFrom:
      kind: AzureStorageAccount
      name: prod-data
      fieldPath: status.outputs.storage_account_name
  accessKey:
    valueFrom:
      kind: AzureStorageAccount
      name: prod-data
      fieldPath: status.outputs.primary_access_key
```

The InfraPipeline resolves the dependency graph, deploys the account, share, and environment first, then registers the storage with the resolved values -- apps and jobs mount it by referencing its `storage_name` output in their volumes.

## Key Configuration

These are the most important decisions when configuring a registration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Protocol** -- exactly one per registration. SMB (`accountName` + `accessKey` together) is the common case and works in every environment; NFS (`nfsServerUrl` alone, the bare `{account}.file.core.windows.net` hostname) requires a VNet-injected environment and a premium FileStorage account, and pairs with NFS_AZURE_FILE volumes.

**Storage name** -- the handle workload volumes carry in their `storageName` field. Name it for the DATA (app-data, ml-models, report-output); the account and share are addressing details.

**Access mode** -- `READ_WRITE` for working storage, `READ_ONLY` for shared configuration and reference data. The mode applies to EVERY volume the registration backs; an environment that needs both postures registers the same share twice under two names.

**Key rotation** -- the SMB `accessKey` is the one in-place update. Referencing the account's `primary_access_key` output (rather than pasting a copied key) keeps the registration correct across out-of-band regenerations. Every other field is fixed at creation, and recreating a registration briefly breaks every volume mount that references it -- plan changes deliberately.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureContainerAppEnvironment** | `containerAppEnvironmentId` | `status.outputs.environment_id` |
| **AzureStorageShare** | `shareName` | `status.outputs.share_name` |
| **AzureStorageAccount** | `accountName` (SMB) | `status.outputs.storage_account_name` |
| **AzureStorageAccount** | `accessKey` (SMB) | `status.outputs.primary_access_key` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `storage_name` | The registration's name on the environment | The volume seam: AzureContainerApp and AzureContainerAppJob volumes reference it in their `storageName` field |
| `storage_id` | Azure Resource Manager ID of the registration | Operational tooling and portal navigation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Persistent app storage** -- an SMB share registered read-write, mounted by one or more apps as their durable working directory. Start from the **SMB Azure Files Share (Read-Write)** preset.

**Producer/consumer split** -- the SAME share registered twice: read-write for the batch job that writes results, read-only for the serving app that reads them -- one share, two registrations, two postures.

**High-throughput NFS** -- a premium FileStorage NFS share for POSIX-style workloads in a VNet-injected environment. Start from the **NFS Azure Files Share** preset.

## Works With

- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- where the registration lives
- [**Azure Storage Share**](/infra-catalog/azure-storage-share) -- the Azure Files share being registered
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- addresses and authenticates the SMB mount
- [**Azure Container App**](/infra-catalog/azure-container-app) -- mounts the registration through AZURE_FILE / NFS_AZURE_FILE volumes
- [**Azure Container App Job**](/infra-catalog/azure-container-app-job) -- persists execution results through the same volume seam
