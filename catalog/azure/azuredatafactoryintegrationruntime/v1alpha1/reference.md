# AzureDataFactoryIntegrationRuntime

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDataFactoryIntegrationRuntimeSpec** defines an Azure Data
Factory integration runtime -- the compute engine a factory's
pipelines, data flows, and copy activities actually run on. The
factory itself stores definitions; an integration runtime is what
executes them.

The engine flavor is declared by which variant block is present:
set exactly ONE of the three blocks below. All flavors share one
factory-scoped name namespace
({factory_id}/integrationRuntimes/{name}).

  - `azure`: the fully managed compute for mapping data flows --
    serverless Spark that Azure provisions on demand and tears down
    after the configured time-to-live.
  - `azure_ssis`: a managed cluster of VMs that runs SQL Server
    Integration Services (SSIS) packages -- the lift-and-shift home
    for existing SSIS projects, with an optional SSISDB catalog,
    custom setup, and virtual network injection.
  - `self_hosted`: a registration for the agent you install on your
    own machines (on-premises or in any network Azure cannot reach
    directly). Azure issues authorization keys the agent uses to
    join; creating the registration is free -- the compute is
    yours.

Billing reality: the azure flavor bills only while data flows run
(plus any warm time-to-live); the SSIS flavor bills per node-hour
from the moment the runtime is STARTED (creating it leaves it
stopped and unbilled); the self-hosted flavor is free on Azure's
side.

## Example

```yaml
# The canonical deep-shape fixture for offline validation: the
# managed data-flow compute at full depth (every azure-variant field
# set), with a literal factory ID so the manifest validates
# standalone. The live lanes use the scenarios/ files instead.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryIntegrationRuntime
metadata:
  name: dataflow-compute
  id: dataflow-compute
  org: planton-oss
  env: e2e
spec:
  dataFactoryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df
  name: dataflow-compute
  description: The managed Spark compute mapping data flows run on, warm-pooled inside the managed virtual network.
  azure:
    region: eastus
    cleanupEnabled: false
    computeType: MemoryOptimized
    coreCount: 16
    timeToLiveMin: 15
    virtualNetworkEnabled: true
    interactiveAuthoringTimeToLiveInMinutes: 30
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dataFactoryId` | `string \| valueFrom` | yes |  | AzureDataFactory (`status.outputs.data_factory_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.azure` | `AzureDataFactoryIntegrationRuntimeAzure` |  |  |  |
| `spec.azure.region` | `string` | yes |  |  |
| `spec.azure.cleanupEnabled` | `bool` |  | `true` |  |
| `spec.azure.computeType` | `string` |  |  |  |
| `spec.azure.coreCount` | `int32` |  |  |  |
| `spec.azure.timeToLiveMin` | `int32` |  |  |  |
| `spec.azure.interactiveAuthoringTimeToLiveInMinutes` | `int32` |  |  |  |
| `spec.azure.virtualNetworkEnabled` | `bool` |  |  |  |
| `spec.azureSsis` | `AzureDataFactoryIntegrationRuntimeAzureSsis` |  |  |  |
| `spec.azureSsis.region` | `string` | yes |  |  |
| `spec.azureSsis.nodeSize` | `string` | yes |  |  |
| `spec.azureSsis.numberOfNodes` | `int32` |  |  |  |
| `spec.azureSsis.maxParallelExecutionsPerNode` | `int32` |  |  |  |
| `spec.azureSsis.edition` | `string` |  |  |  |
| `spec.azureSsis.licenseType` | `string` |  |  |  |
| `spec.azureSsis.credentialName` | `string` |  |  |  |
| `spec.azureSsis.catalogInfo` | `AzureDataFactoryIntegrationRuntimeSsisCatalogInfo` |  |  |  |
| `spec.azureSsis.catalogInfo.serverEndpoint` | `string` | yes |  |  |
| `spec.azureSsis.catalogInfo.administratorLogin` | `string` |  |  |  |
| `spec.azureSsis.catalogInfo.administratorPassword` | `string` (sensitive) |  |  |  |
| `spec.azureSsis.catalogInfo.pricingTier` | `string` |  |  |  |
| `spec.azureSsis.catalogInfo.elasticPoolName` | `string` |  |  |  |
| `spec.azureSsis.catalogInfo.dualStandbyPairName` | `string` |  |  |  |
| `spec.azureSsis.customSetupScript` | `AzureDataFactoryIntegrationRuntimeSsisCustomSetupScript` |  |  |  |
| `spec.azureSsis.customSetupScript.blobContainerUri` | `string` | yes |  |  |
| `spec.azureSsis.customSetupScript.sasToken` | `string` (sensitive) | yes |  |  |
| `spec.azureSsis.expressCustomSetup` | `AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup` |  |  |  |
| `spec.azureSsis.expressCustomSetup.environment` | `map<string, string>` |  |  |  |
| `spec.azureSsis.expressCustomSetup.powershellVersion` | `string` |  |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey` | `[]AzureDataFactoryIntegrationRuntimeSsisCommandKey` |  |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].targetName` | `string` | yes |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].userName` | `string` | yes |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].password` | `string` (sensitive) |  |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword` | `AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference` |  |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.parameters` | `map<string, string>` |  |  |  |
| `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.secretVersion` | `string` |  |  |  |
| `spec.azureSsis.expressCustomSetup.component` | `[]AzureDataFactoryIntegrationRuntimeSsisComponent` |  |  |  |
| `spec.azureSsis.expressCustomSetup.component[].name` | `string` | yes |  |  |
| `spec.azureSsis.expressCustomSetup.component[].license` | `string` (sensitive) |  |  |  |
| `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense` | `AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference` |  |  |  |
| `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.secretName` | `string` | yes |  |  |
| `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.parameters` | `map<string, string>` |  |  |  |
| `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.secretVersion` | `string` |  |  |  |
| `spec.azureSsis.expressVnetIntegration` | `AzureDataFactoryIntegrationRuntimeSsisExpressVnetIntegration` |  |  |  |
| `spec.azureSsis.expressVnetIntegration.subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.azureSsis.vnetIntegration` | `AzureDataFactoryIntegrationRuntimeSsisVnetIntegration` |  |  |  |
| `spec.azureSsis.vnetIntegration.vnetId` | `string \| valueFrom` |  |  | AzureVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.azureSsis.vnetIntegration.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.azureSsis.vnetIntegration.subnetName` | `string` |  |  |  |
| `spec.azureSsis.vnetIntegration.publicIps` | `[]string \| valueFrom` |  |  | AzurePublicIp (`status.outputs.public_ip_id`) |
| `spec.azureSsis.packageStore` | `[]AzureDataFactoryIntegrationRuntimeSsisPackageStore` |  |  |  |
| `spec.azureSsis.packageStore[].name` | `string` | yes |  |  |
| `spec.azureSsis.packageStore[].linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureSsis.copyComputeScale` | `AzureDataFactoryIntegrationRuntimeSsisCopyComputeScale` |  |  |  |
| `spec.azureSsis.copyComputeScale.dataIntegrationUnit` | `int32` |  |  |  |
| `spec.azureSsis.copyComputeScale.timeToLive` | `int32` |  |  |  |
| `spec.azureSsis.pipelineExternalComputeScale` | `AzureDataFactoryIntegrationRuntimeSsisPipelineExternalComputeScale` |  |  |  |
| `spec.azureSsis.pipelineExternalComputeScale.numberOfExternalNodes` | `int32` |  |  |  |
| `spec.azureSsis.pipelineExternalComputeScale.numberOfPipelineNodes` | `int32` |  |  |  |
| `spec.azureSsis.pipelineExternalComputeScale.timeToLive` | `int32` |  |  |  |
| `spec.azureSsis.proxy` | `AzureDataFactoryIntegrationRuntimeSsisProxy` |  |  |  |
| `spec.azureSsis.proxy.selfHostedIntegrationRuntimeName` | `string \| valueFrom` | yes |  | AzureDataFactoryIntegrationRuntime (`status.outputs.integration_runtime_name`) |
| `spec.azureSsis.proxy.stagingStorageLinkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureSsis.proxy.path` | `string` |  |  |  |
| `spec.selfHosted` | `AzureDataFactoryIntegrationRuntimeSelfHosted` |  |  |  |
| `spec.selfHosted.rbacAuthorization` | `AzureDataFactoryIntegrationRuntimeSelfHostedRbacAuthorization` |  |  |  |
| `spec.selfHosted.rbacAuthorization.resourceId` | `string \| valueFrom` | yes |  | AzureDataFactoryIntegrationRuntime (`status.outputs.integration_runtime_id`) |
| `spec.selfHosted.selfContainedInteractiveAuthoringEnabled` | `bool` |  |  |  |

## Field Details

### spec.dataFactoryId

`string | valueFrom` · required

The Data Factory the integration runtime lives in, by ARM ID.
Can be a literal string or a reference to an AzureDataFactory
output.

**ForceNew**: changing this destroys and recreates the runtime.

- references: AzureDataFactory (`status.outputs.data_factory_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactory, name: <that resource's name>, fieldPath: status.outputs.data_factory_id}} -- a bare string does not parse

### spec.name

`string` · required

The integration runtime's name -- unique within the factory
across all three flavors. Letters, numbers, and dashes; must
start and end with a letter or number; no consecutive dashes.
The managed flavors (azure, azure_ssis) additionally require at
least 3 characters -- the two format rules are enforced by the
message rules below, each transcribed from Azure's own validator
for that flavor.

**ForceNew**: changing this destroys and recreates the runtime.

- rule: {"required":true}

### spec.description

`string`

A human-readable description of what the runtime is for.

### spec.azure

`AzureDataFactoryIntegrationRuntimeAzure`

The managed data-flow compute engine.

- rule: interactive_authoring_time_to_live_in_minutes requires virtual_network_enabled to be true

### spec.azure.region

`string` · required

The Azure region the compute is provisioned in, e.g. "eastus" --
or the literal "AutoResolve" to run each data flow in the region
closest to its data (the Data Factory Studio default).

**ForceNew**: changing this destroys and recreates the runtime.

- rule: {"required":true}

### spec.azure.cleanupEnabled

`bool` · optional (explicit presence)

Whether the cluster is torn down immediately after every data
flow run. Unspecified applies true (Azure's default). Set false
together with time_to_live_min to keep a warm pool between runs.

- default: `true`

### spec.azure.computeType

`string`

The compute profile of the Spark cluster -- omit for "General".

- rule: {"string":{"in":["","General","ComputeOptimized","MemoryOptimized"]}}

### spec.azure.coreCount

`int32`

The cluster's core count -- omit for 8 (Azure's smallest data
flow cluster).

- rule: core_count must be one of 8, 16, 32, 48, 80, 144, or 272

### spec.azure.timeToLiveMin

`int32`

Minutes a cluster stays warm after a data flow run before the
cleanup -- omit for 0 (tear down immediately). Warm minutes bill
like run minutes.

- rule: {"int32":{"gte":0}}

### spec.azure.interactiveAuthoringTimeToLiveInMinutes

`int32`

Minutes an interactive authoring (debug) session stays alive
before auto-terminating -- one of 10, 30, 60, or 120; omit to
leave interactive authoring disabled. Requires
virtual_network_enabled (Azure's own rule). Azure applies this
through a separate enable-interactive-authoring operation after
the runtime is online, and a live debug session bills while
enabled -- both engines handle the extra call.

- rule: interactive_authoring_time_to_live_in_minutes must be one of 10, 30, 60, or 120

### spec.azure.virtualNetworkEnabled

`bool`

Whether the compute joins the factory's managed virtual network
(so data flows can reach private endpoints). The factory must
have its managed virtual network enabled, or the deploy fails
with Azure's own error.

**ForceNew**: changing this destroys and recreates the runtime.

### spec.azureSsis

`AzureDataFactoryIntegrationRuntimeAzureSsis`

The managed SSIS package runtime.

### spec.azureSsis.region

`string` · required

The Azure region the SSIS nodes run in, e.g. "eastus".

**ForceNew**: changing this destroys and recreates the runtime.

- rule: {"required":true}

### spec.azureSsis.nodeSize

`string` · required

The VM size of each node (Azure's fixed menu for SSIS runtimes).

- rule: {"required":true,"string":{"in":["Standard_D2_v3","Standard_D4_v3","Standard_D8_v3","Standard_D16_v3","Standard_D32_v3","Standard_D64_v3","Standard_E2_v3","Standard_E4_v3","Standard_E8_v3","Standard_E16_v3","Standard_E32_v3","Standard_E64_v3","Standard_D1_v2","Standard_D2_v2","Standard_D3_v2","Standard_D4_v2","Standard_A4_v2","Standard_A8_v2"]}}

### spec.azureSsis.numberOfNodes

`int32`

How many nodes the cluster runs (1-10) -- omit for 1.

- rule: number_of_nodes must be between 1 and 10

### spec.azureSsis.maxParallelExecutionsPerNode

`int32`

How many packages one node executes in parallel (1-16) -- omit
for 1.

- rule: max_parallel_executions_per_node must be between 1 and 16

### spec.azureSsis.edition

`string`

The SSIS edition of the nodes -- omit for "Standard".
"Enterprise" unlocks the Enterprise-only SSIS features.

- rule: {"string":{"in":["","Standard","Enterprise"]}}

### spec.azureSsis.licenseType

`string`

How the SQL Server licenses on the nodes are paid for -- omit
for "LicenseIncluded". "BasePrice" applies Azure Hybrid Benefit
(bring your own license).

- rule: {"string":{"in":["","LicenseIncluded","BasePrice"]}}

### spec.azureSsis.credentialName

`string`

The name of a factory credential (a user-assigned managed
identity registered on the AzureDataFactory itself, under its
`credentials`) the runtime authenticates with -- e.g. toward the
SSISDB catalog server. Omit to use the factory's system identity.

### spec.azureSsis.catalogInfo

`AzureDataFactoryIntegrationRuntimeSsisCatalogInfo`

Hosts the SSISDB catalog (the project/package store SSIS Studio
deploys to) on an Azure SQL server. Omit to run packages without
a catalog (files/package stores only).

- rule: pricing_tier and elastic_pool_name are mutually exclusive -- SSISDB lands either on a tier or in a pool

### spec.azureSsis.catalogInfo.serverEndpoint

`string` · required

The Azure SQL server's endpoint, e.g.
myserver.database.windows.net.

- rule: {"required":true}

### spec.azureSsis.catalogInfo.administratorLogin

`string`

The server's SQL administrator login -- omit to authenticate as
the factory's managed identity (grant it access on the server).

### spec.azureSsis.catalogInfo.administratorPassword

`string` · sensitive

The administrator's password. Stored by Azure as a hidden secure
string. SECRET. Omit when authenticating with the managed
identity.

### spec.azureSsis.catalogInfo.pricingTier

`string`

The service tier of the SSISDB database Azure creates
(DTU or vCore SKU name) -- omit for Azure's default. Mutually
exclusive with elastic_pool_name.

- rule: {"string":{"in":["","Basic","S0","S1","S2","S3","S4","S6","S7","S9","S12","P1","P2","P4","P6","P11","P15","GP_S_Gen5_1","GP_S_Gen5_2","GP_S_Gen5_4","GP_S_Gen5_6","GP_S_Gen5_8","GP_S_Gen5_10","GP_S_Gen5_12","GP_S_Gen5_14","GP_S_Gen5_16","GP_S_Gen5_18","GP_S_Gen5_20","GP_S_Gen5_24","GP_S_Gen5_32","GP_S_Gen5_40","GP_Gen5_2","GP_Gen5_4","GP_Gen5_6","GP_Gen5_8","GP_Gen5_10","GP_Gen5_12","GP_Gen5_14","GP_Gen5_16","GP_Gen5_18","GP_Gen5_20","GP_Gen5_24","GP_Gen5_32","GP_Gen5_40","GP_Gen5_80","BC_Gen5_2","BC_Gen5_4","BC_Gen5_6","BC_Gen5_8","BC_Gen5_10","BC_Gen5_12","BC_Gen5_14","BC_Gen5_16","BC_Gen5_18","BC_Gen5_20","BC_Gen5_24","BC_Gen5_32","BC_Gen5_40","BC_Gen5_80","HS_Gen5_2","HS_Gen5_4","HS_Gen5_6","HS_Gen5_8","HS_Gen5_10","HS_Gen5_12","HS_Gen5_14","HS_Gen5_16","HS_Gen5_18","HS_Gen5_20","HS_Gen5_24","HS_Gen5_32","HS_Gen5_40","HS_Gen5_80"]}}

### spec.azureSsis.catalogInfo.elasticPoolName

`string`

The elastic pool on the server to create SSISDB in -- the
alternative to picking a pricing_tier. Mutually exclusive with
pricing_tier.

### spec.azureSsis.catalogInfo.dualStandbyPairName

`string`

The name of the dual-standby pair for SSISDB geo-failover -- set
when the catalog participates in a failover group.

### spec.azureSsis.customSetupScript

`AzureDataFactoryIntegrationRuntimeSsisCustomSetupScript`

A custom setup script container: a blob container holding
main.cmd plus installers, executed on every node at start. The
slower, fully general alternative to express_custom_setup.

### spec.azureSsis.customSetupScript.blobContainerUri

`string` · required

The blob container's URI holding main.cmd and its installers.

- rule: {"required":true}

### spec.azureSsis.customSetupScript.sasToken

`string` · required · sensitive

A SAS token granting read access to the container. Stored by
Azure as a hidden secure string. SECRET.

- rule: {"required":true}

### spec.azureSsis.expressCustomSetup

`AzureDataFactoryIntegrationRuntimeSsisExpressCustomSetup`

Express custom setup: environment variables, a PowerShell
version, licensed components, and Windows credential-manager
entries applied to every node at start -- without a setup script
container.

- rule: Declare at least one of environment, powershell_version, command_key, or component

### spec.azureSsis.expressCustomSetup.environment

`map<string, string>`

Environment variables set on every node, keyed by variable name.

### spec.azureSsis.expressCustomSetup.powershellVersion

`string`

The Azure PowerShell version installed on every node, e.g.
"7.2.1".

### spec.azureSsis.expressCustomSetup.commandKey

`[]AzureDataFactoryIntegrationRuntimeSsisCommandKey`

Windows credential-manager entries (cmdkey) created on every
node, so packages can authenticate to file shares and servers.

### spec.azureSsis.expressCustomSetup.commandKey[].targetName

`string` · required

The target the credential applies to (server or share name).

- rule: {"required":true}

### spec.azureSsis.expressCustomSetup.commandKey[].userName

`string` · required

The user name.

- rule: {"required":true}

### spec.azureSsis.expressCustomSetup.commandKey[].password

`string` · sensitive

The password, inline. Stored by Azure as a hidden secure string.
SECRET. Prefer key_vault_password to keep the value out of
manifests.

### spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword

`AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference`

The password, read from Key Vault through a Key Vault linked
service -- the secretless alternative to password.

### spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.secretName

`string` · required

The secret's name in the vault.

- rule: {"required":true}

### spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.parameters

`map<string, string>`

Values for the linked service's parameters, keyed by parameter
name.

### spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.secretVersion

`string`

The secret's version -- omit for the latest.

### spec.azureSsis.expressCustomSetup.component

`[]AzureDataFactoryIntegrationRuntimeSsisComponent`

Licensed third-party components installed on every node.

### spec.azureSsis.expressCustomSetup.component[].name

`string` · required

The component's name, as its installer registers it.

- rule: {"required":true}

### spec.azureSsis.expressCustomSetup.component[].license

`string` · sensitive

The license key, inline. Stored by Azure as a hidden secure
string. SECRET. Prefer key_vault_license to keep the value out
of manifests.

### spec.azureSsis.expressCustomSetup.component[].keyVaultLicense

`AzureDataFactoryIntegrationRuntimeSsisKeyVaultSecretReference`

The license key, read from Key Vault through a Key Vault linked
service -- the secretless alternative to license.

### spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.secretName

`string` · required

The secret's name in the vault.

- rule: {"required":true}

### spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.parameters

`map<string, string>`

Values for the linked service's parameters, keyed by parameter
name.

### spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.secretVersion

`string`

The secret's version -- omit for the latest.

### spec.azureSsis.expressVnetIntegration

`AzureDataFactoryIntegrationRuntimeSsisExpressVnetIntegration`

Express virtual network injection: the runtime reaches the
subnet through Azure's injection without you delegating the
subnet to Data Factory. The lighter alternative to
vnet_integration (standard injection).

### spec.azureSsis.expressVnetIntegration.subnetId

`string | valueFrom` · required

The subnet the runtime reaches, by ARM ID -- defaults to
referencing an AzureSubnet's subnet_id output.

The runtime is an ARM child of its factory and lives there; the
subnet is where its nodes attach, so on a diagram the reference is
access, not placement -- the AKS node pool's rule.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.azureSsis.vnetIntegration

`AzureDataFactoryIntegrationRuntimeSsisVnetIntegration`

Standard virtual network injection: the nodes join a subnet you
own (address the network by ID + subnet name, or the subnet by
ID directly).

- rule: Set exactly one of vnet_id (with subnet_name) or subnet_id
- rule: subnet_name is required with vnet_id (and is only meaningful there)
- rule: public_ips takes exactly two addresses (or none, letting Azure assign them)

### spec.azureSsis.vnetIntegration.vnetId

`string | valueFrom`

The virtual network, by ARM ID -- pairs with subnet_name.
Defaults to referencing an AzureVirtualNetwork's
virtual_network_id output.

The runtime is an ARM child of its factory and lives there; the
network is where its nodes attach, so on a diagram the reference is
access, not placement -- the AKS node pool's rule.

- references: AzureVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.azureSsis.vnetIntegration.subnetId

`string | valueFrom`

The subnet, by ARM ID -- the direct alternative to vnet_id +
subnet_name. Defaults to referencing an AzureSubnet's subnet_id
output.

Access, not placement, for the same reason as vnet_id above.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.azureSsis.vnetIntegration.subnetName

`string`

The subnet's name inside vnet_id -- required with vnet_id,
meaningless with subnet_id.

### spec.azureSsis.vnetIntegration.publicIps

`[]string | valueFrom`

Exactly TWO standard static public IPs (by ARM ID) the runtime
presents as its outbound addresses -- omit to let Azure assign
them. Each defaults to referencing an AzurePublicIp's
public_ip_id output.

- references: AzurePublicIp (`status.outputs.public_ip_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIp, name: <that resource's name>, fieldPath: status.outputs.public_ip_id}} -- a bare string does not parse

### spec.azureSsis.packageStore

`[]AzureDataFactoryIntegrationRuntimeSsisPackageStore`

Package stores: named file/MSDB locations (each reached through
a linked service) that SSIS Studio can deploy packages to,
instead of -- or alongside -- the SSISDB catalog.

### spec.azureSsis.packageStore[].name

`string` · required

The store's display name in SSIS Studio.

- rule: {"required":true}

### spec.azureSsis.packageStore[].linkedServiceName

`string | valueFrom` · required

The linked service carrying the store's location (an Azure file
share or SQL MSDB connection), by name -- defaults to
referencing an AzureDataFactoryLinkedService's
linked_service_name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureSsis.copyComputeScale

`AzureDataFactoryIntegrationRuntimeSsisCopyComputeScale`

Scales the copy activities that run ON this runtime -- data
integration units and their warm time-to-live.

### spec.azureSsis.copyComputeScale.dataIntegrationUnit

`int32`

Data integration units powering each copy activity -- a multiple
of 4 between 4 and 256. Omit for Azure's default.

- rule: data_integration_unit must be a multiple of 4 between 4 and 256

### spec.azureSsis.copyComputeScale.timeToLive

`int32`

Minutes the copy compute stays warm between activities -- at
least 5. Omit for Azure's default.

- rule: time_to_live must be at least 5 minutes

### spec.azureSsis.pipelineExternalComputeScale

`AzureDataFactoryIntegrationRuntimeSsisPipelineExternalComputeScale`

Scales the pipeline/external activities that run ON this runtime
-- node counts and their warm time-to-live.

### spec.azureSsis.pipelineExternalComputeScale.numberOfExternalNodes

`int32`

Nodes reserved for external activities (1-10). Omit for Azure's
default. Note: Azure's read API does not report this value back
(it mirrors number_of_pipeline_nodes instead -- a provider read
seam both engines inherit), so a re-read after deploy shows the
pipeline-node count here.

- rule: number_of_external_nodes must be between 1 and 10

### spec.azureSsis.pipelineExternalComputeScale.numberOfPipelineNodes

`int32`

Nodes reserved for pipeline activities (1-10). Omit for Azure's
default.

- rule: number_of_pipeline_nodes must be between 1 and 10

### spec.azureSsis.pipelineExternalComputeScale.timeToLive

`int32`

Minutes the pipeline compute stays warm between activities -- at
least 5. Omit for Azure's default.

- rule: time_to_live must be at least 5 minutes

### spec.azureSsis.proxy

`AzureDataFactoryIntegrationRuntimeSsisProxy`

Routes the runtime's on-premises data access through a
self-hosted integration runtime (the proxy), staging data
through a storage linked service.

### spec.azureSsis.proxy.selfHostedIntegrationRuntimeName

`string | valueFrom` · required

The self-hosted integration runtime acting as the proxy, by name
-- defaults to referencing another
AzureDataFactoryIntegrationRuntime's integration_runtime_name
output (a self_hosted one, in the same factory).

- references: AzureDataFactoryIntegrationRuntime (`status.outputs.integration_runtime_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryIntegrationRuntime, name: <that resource's name>, fieldPath: status.outputs.integration_runtime_name}} -- a bare string does not parse

### spec.azureSsis.proxy.stagingStorageLinkedServiceName

`string | valueFrom` · required

The storage linked service the proxy stages data through, by
name -- defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureSsis.proxy.path

`string`

The folder path in the staging store -- omit for the store's
root.

### spec.selfHosted

`AzureDataFactoryIntegrationRuntimeSelfHosted`

The self-hosted agent registration.

### spec.selfHosted.rbacAuthorization

`AzureDataFactoryIntegrationRuntimeSelfHostedRbacAuthorization`

Shares ANOTHER factory's self-hosted runtime instead of
registering new compute: the runtime this one links to, granted
through Azure RBAC. Omit for a primary (own-compute)
registration. A linked runtime issues no authorization keys of
its own.

**ForceNew**: changing this destroys and recreates the
registration.

### spec.selfHosted.rbacAuthorization.resourceId

`string | valueFrom` · required

The PRIMARY self-hosted integration runtime being shared, by ARM
ID ({factory_id}/integrationRuntimes/{name}) -- defaults to
referencing another AzureDataFactoryIntegrationRuntime's
integration_runtime_id output.

- references: AzureDataFactoryIntegrationRuntime (`status.outputs.integration_runtime_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryIntegrationRuntime, name: <that resource's name>, fieldPath: status.outputs.integration_runtime_id}} -- a bare string does not parse

### spec.selfHosted.selfContainedInteractiveAuthoringEnabled

`bool`

Whether interactive authoring runs self-contained on the
runtime's own nodes (instead of relaying through Azure) --
needed when the nodes cannot reach Azure's relay endpoints.
Unspecified applies false.

## Validation Rules

- `azure_data_factory_integration_runtime_name_format_managed`: Managed runtime names need at least 3 characters -- letters, numbers, and dashes only, starting and ending with a letter or number, no consecutive dashes
- `azure_data_factory_integration_runtime_name_format_self_hosted`: Self-hosted runtime names use letters, numbers, and dashes, starting and ending with a letter or number, with no consecutive dashes

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactoryIntegrationRuntime, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.integration_runtime_id` | `string` | The integration runtime's Azure Resource Manager ID ({factory_id}/integrationRuntimes/{name}) -- the same ID shape for all three flavors. |
| `status.outputs.integration_runtime_name` | `string` | The integration runtime's name -- what linked services, data flow activities, and the SSIS proxy resolve against. |
| `status.outputs.primary_authorization_key` | `string` | The primary authorization key a self-hosted agent uses to join this runtime. Populated only for a PRIMARY self_hosted runtime (Azure issues no keys for the managed flavors or for a linked registration). SECRET -- Azure returns it readable, so the catalog treats it as sensitive even though the provider does not. |
| `status.outputs.secondary_authorization_key` | `string` | The secondary authorization key (for key rotation). Populated only for a PRIMARY self_hosted runtime. SECRET. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataFactoryId` | AzureDataFactory | `status.outputs.data_factory_id` |
| `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureSsis.expressVnetIntegration.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.azureSsis.vnetIntegration.vnetId` | AzureVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.azureSsis.vnetIntegration.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.azureSsis.vnetIntegration.publicIps` | AzurePublicIp | `status.outputs.public_ip_id` |
| `spec.azureSsis.packageStore[].linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureSsis.proxy.selfHostedIntegrationRuntimeName` | AzureDataFactoryIntegrationRuntime | `status.outputs.integration_runtime_name` |
| `spec.azureSsis.proxy.stagingStorageLinkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.selfHosted.rbacAuthorization.resourceId` | AzureDataFactoryIntegrationRuntime | `status.outputs.integration_runtime_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.proxy.selfHostedIntegrationRuntimeName` | `status.outputs.integration_runtime_name` |
| AzureDataFactoryIntegrationRuntime | `spec.selfHosted.rbacAuthorization.resourceId` | `status.outputs.integration_runtime_id` |
| AzureDataFactoryLinkedService | `spec.integrationRuntimeName` | `status.outputs.integration_runtime_name` |

## See Also

- [Overview](../README.md)
