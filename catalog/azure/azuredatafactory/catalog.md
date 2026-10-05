# Azure Data Factory

Deploys an Azure Data Factory -- the workspace every other Data Factory resource lives inside: pipelines, data flows, linked services, datasets, triggers, and integration runtimes are all created against the factory's ARM ID. The factory carries the workspace-level posture: its managed identity, an optional git repository binding, workspace-wide global parameters, customer-managed-key encryption, named credentials, a managed virtual network, and managed private endpoints for private egress. The factory itself is near-free at rest; billing follows pipeline activity.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data Factory** -- the workspace itself: managed identity, optional GitHub or Azure DevOps repository binding, global parameters, public-network posture, optional Purview connection, and inline customer-managed-key encryption
- **Named credentials** (created only when `userManagedIdentityCredentials` or `servicePrincipalCredentials` are set) -- one per entry, wrapping a user-assigned identity or a service principal whose key lives in Key Vault; linked services reference them by name
- **Managed private endpoints** (created only when `managedPrivateEndpoints` are set) -- private egress from the factory's managed virtual network to data stores; requires the managed VNet

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** -- referenced through `resourceGroup` as a literal name or an AzureResourceGroup ValueFromRef.
- **A Key Vault key and unwrap identity** (only for CMK encryption) -- an AzureKeyVaultKey (prefer its versionless ID so rotation propagates) and an AzureUserAssignedIdentity holding get/unwrap/wrap on the vault BEFORE create, attached in the identity block. Azure validates both at deploy time.
- **A globally unique factory name** -- Data Factory names are unique across all of Azure (the name becomes part of the factory's URL); prefix with your org, or a taken name fails at deploy time.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Data Platform Workspace** or **Private ETL Workspace** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactory
metadata:
  name: data-platform
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  name: acme-data-platform
  region: eastus
  identity:
    type: SYSTEM_ASSIGNED
  managedVirtualNetworkEnabled: true
```

```shell
planton apply -f data-factory.yaml
```

This creates a factory with a system-assigned identity and the managed virtual network enabled -- near-free at rest, ready for pipelines and private endpoints to onboard against it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the resource group -- and, for CMK, the key and unwrap identity -- by reference:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: data-rg
      fieldPath: status.outputs.resource_group_name
  name: acme-data-platform
  region: eastus
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - valueFrom:
          kind: AzureUserAssignedIdentity
          name: adf-cmk-identity
          fieldPath: status.outputs.identity_id
  customerManagedKey:
    keyVaultKeyId:
      valueFrom:
        kind: AzureKeyVaultKey
        name: adf-cmk
        fieldPath: status.outputs.versionless_id
    userAssignedIdentityId:
      valueFrom:
        kind: AzureUserAssignedIdentity
        name: adf-cmk-identity
        fieldPath: status.outputs.identity_id
  managedVirtualNetworkEnabled: true
```

The InfraPipeline resolves the dependency graph, deploys the resource group, identity, and key first, then provisions the factory with the resolved references.

## Key Configuration

These are the most important decisions when configuring a Data Factory. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Decide the managed-VNet question before you create** -- Enabling `managedVirtualNetworkEnabled` is an in-place update, but disabling it replaces the factory -- and a replaced factory drops every pipeline, dataset, and linked service inside it. If there is any chance the factory will need private egress (managed private endpoints require the managed VNet), enable it at creation: the network is free and changes nothing until endpoints use it.

**One factory per environment, not per team** -- A factory at rest costs almost nothing; billing follows pipeline activity (integration runtime hours, data movement volume). Workspace-level settings -- identity, credentials, network posture -- are shared by everything inside, so treat the factory manifest as platform-owned and let teams own their pipelines.

**Private endpoints complete on the other side** -- A managed private endpoint reaches Succeeded while its connection on the TARGET resource is still Pending: the deploy going green does not mean traffic flows. Approval lives on the target (storage account, SQL server) under Networking, usually owned by a different team -- wire it into the onboarding runbook, or pipelines fail at runtime long after the factory deployed cleanly. Each endpoint entry is create-only: any change replaces that endpoint, siblings untouched.

**Removing the git block does not detach the repo** -- The repository binding (`githubConfiguration` or `vstsConfiguration`, at most one) is applied through a side-channel call after the factory exists, and the provider never calls a repo-clear API: deleting the block from the manifest leaves the factory bound. Detach deliberately in the Data Factory Studio. With a repo bound, the collaboration branch is the source of truth and "Publish" is how changes reach the live factory.

**CMK is a one-way door** -- Customer-managed-key encryption can be enabled at create but never removed; Azure has no decrypt path back to service-managed keys. Prefer the versionless key reference so rotation propagates without touching the factory, and remember the unwrap identity must hold vault permissions before create -- Azure validates at deploy time, not plan time.

**Credentials share one namespace** -- User-managed-identity credentials and service-principal credentials share a single name namespace under the factory; linked services reference them by name, so a rename replaces the credential and breaks every linked service pointing at it. Service-principal keys are read through a Key Vault linked service -- the secret itself never enters this spec.

**Public network access is a posture, not a detail** -- `publicNetworkEnabled: false` makes managed private endpoints the factory's only data path. Pair it with identity-based authentication on the stores; flipping it later is an in-place update, but the endpoint approvals must already be in place or everything stops.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** (optional) | `identity.identityIds[]`, `customerManagedKey.userAssignedIdentityId`, `userManagedIdentityCredentials[].identityId` | `status.outputs.identity_id` |
| **AzureKeyVaultKey** (CMK) | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `data_factory_id` | The factory's ARM ID | The `dataFactoryId` every in-factory kind references: pipelines, data flows, datasets, linked services, triggers, integration runtimes |
| `identity_principal_id` | The system-assigned identity's principal ID (empty without one) | AzureRoleAssignment grants on the data stores the factory reads and writes |
| `credential_ids` | ARM IDs of the named credentials, keyed by name | Correlating credentials referenced by linked services |
| `managed_private_endpoint_ids` | ARM IDs of the managed private endpoints, keyed by name | Locating the connections awaiting approval on their targets |

The other output, `data_factory_name`, echoes the factory's name back for reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Environment data platform** -- one factory per environment with a system-assigned identity and the managed VNet enabled up front, pipelines onboarding against its `data_factory_id`. Bind git later (an in-place update) once the authoring repo exists. Start from the **Data Platform Workspace** preset.

**Locked-down private ETL** -- public access off, integration running inside the managed VNet, one managed private endpoint per data store, identity-based auth on the stores. Remember each endpoint's approval on the target side before the first pipeline runs. Start from the **Private ETL Workspace** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the factory lives in
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- attached identities for CMK unwrapping and named credentials
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed encryption key
- [**Azure Data Factory Pipeline**](/infra-catalog/azure-data-factory-pipeline) -- the pipelines created against the factory's `data_factory_id`
- [**Azure Data Factory Linked Service**](/infra-catalog/azure-data-factory-linked-service) -- connections to data stores, authenticating as the factory's identity or named credentials
- [**Azure Data Factory Dataset**](/infra-catalog/azure-data-factory-dataset) -- named views over linked-service data
- [**Azure Data Factory Data Flow**](/infra-catalog/azure-data-factory-data-flow) -- visually-authored transformations executed by pipelines
- [**Azure Data Factory Trigger**](/infra-catalog/azure-data-factory-trigger) -- schedules and events that start pipeline runs
- [**Azure Data Factory Integration Runtime**](/infra-catalog/azure-data-factory-integration-runtime) -- the compute pipelines and data flows execute on
