# Azure Data Factory Integration Runtime

Deploys one integration runtime inside an Azure Data Factory -- the compute engine the factory's pipelines, data flows, and copy activities actually run on, in one of three flavors: the managed data-flow compute (serverless Spark), the managed SSIS package runtime (a cluster of Azure-managed VMs), or the self-hosted agent registration for machines Azure cannot reach directly. The three flavors bill three different ways: data-flow compute per vCore-hour while clusters run, SSIS per node-hour from start to stop (created STOPPED and unbilled), and self-hosted free on Azure's side.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one integration runtime of the flavor the spec's variant block declares:

- **Azure (data-flow compute)** -- serverless Spark Azure provisions when a mapping data flow runs, sized by compute type and core count, optionally kept warm between runs and joined to the factory's managed virtual network
- **Azure-SSIS** -- a managed cluster of VMs that runs SQL Server Integration Services packages, with an optional SSISDB catalog on your Azure SQL server, node custom setup (script container or express form), virtual network injection (standard or express), package stores, copy/pipeline compute scaling, and an on-premises proxy through a self-hosted runtime
- **Self-hosted** -- the registration for the agent you install on your own machines; Azure issues two authorization keys the agent joins with (surfaced as sensitive outputs), or an RBAC-authorized link to another factory's primary runtime

All three flavors share one factory-scoped name namespace (`{factory_id}/integrationRuntimes/{name}`).

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Data Factory** -- the runtime lives in a factory; reference an AzureDataFactory's `data_factory_id` output or provide the ARM ID directly.
- **For the azure flavor's virtual network switch** -- the factory must be deployed with its managed virtual network enabled; Azure rejects the runtime otherwise.

### Azure Subscription

- **An Azure SQL server** (only for the SSIS catalog) -- `catalogInfo.serverEndpoint` names the server SSISDB is created on when the runtime first starts; grant the factory's managed identity access when omitting the SQL administrator login.
- **A subnet** (only for SSIS virtual network injection) -- standard injection joins the nodes to a subnet you own; express injection reaches it without delegating the subnet to Data Factory.
- **A self-hosted registration needs an agent** -- creating it issues keys but moves no data until you install the integration runtime agent on your machines and hand it a key.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory Integration Runtime**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Data Flow Compute**, **Self-Hosted Bridge**, or **SSIS Runtime** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryIntegrationRuntime
metadata:
  name: dataflow-compute
  org: acme-corp
  env: prod
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  name: dataflow-compute
  azure:
    region: eastus
    timeToLiveMin: 10
```

```shell
planton apply -f data-factory-integration-runtime.yaml
```

This creates the managed data-flow compute in eastus with Azure's smallest cluster (General profile, 8 cores) and a 10-minute warm pool between runs -- billing only while a cluster is up. An Infra Job tracks the provisioning in real time.

For the SSIS flavor's secret-bearing fields -- the catalog's `administratorPassword`, the setup script's `sasToken`, cmdkey passwords, and component licenses -- reference managed secrets as `$secret/<slug>` instead of pasting values, or use the Key Vault reference alternatives where the spec offers them.

### InfraChart

When deploying the factory and its runtimes as one chart, ValueFromRef wires the runtime to the factory -- and an SSIS runtime to its subnet or its self-hosted proxy -- deployed in the same InfraPipeline:

```yaml
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  azureSsis:
    region: eastus
    nodeSize: Standard_D4_v3
    expressVnetIntegration:
      subnetId:
        valueFrom:
          kind: AzureSubnet
          name: ssis-subnet
          fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph -- factory and network first, then this runtime.

## Key Configuration

These are the most important decisions when configuring an integration runtime. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The flavor decides the billing model** -- data-flow compute bills per vCore-hour for every run plus every warm minute `timeToLiveMin` buys; the SSIS runtime bills per node-hour from START to STOP (this component only creates it, stopped -- starting is an operational action in Data Factory Studio or via the Data Factory API, and a runtime left started over a weekend is the classic surprise bill); self-hosted is free on Azure's side because the compute is yours.

**One name, one runtime** -- all three flavors live in one factory-scoped namespace, so changing the variant block replaces the object at the same ARM address, and every linked service or activity referencing the name picks up the new engine immediately. `name` and `dataFactoryId` are ForceNew. Rename rather than reshape when anything still depends on the old flavor.

**Warm pools trade money for latency** -- a data flow against a cold runtime pays several minutes of cluster startup. `timeToLiveMin` keeps the cluster warm between runs and `cleanupEnabled: false` preserves the pool, so back-to-back flows start in seconds. Set the TTL to your inter-run gap, not longer: warm minutes bill exactly like run minutes, and `AutoResolve` regions cannot share warm pools across regions.

**The managed virtual network switch is a create-time decision** -- `virtualNetworkEnabled` requires the factory's managed virtual network, and it is ForceNew: flipping it later replaces the runtime. A data flow that must reach a private endpoint needs the whole chain -- factory managed VNet, runtime inside it, managed private endpoint -- designed together. Interactive authoring (`interactiveAuthoringTimeToLiveInMinutes`) is only valid inside the managed virtual network, and a live debug session bills while enabled.

**Self-hosted authorization keys are full join credentials** -- Azure returns them readable; the component surfaces them as sensitive outputs. Any machine holding a key can register as YOUR runtime and see the data flowing through it. Wire keys into installers by reference, rotate with the secondary key (it exists exactly so agents migrate before the primary rotates), and never paste them into manifests. A linked registration (`rbacAuthorization`) issues no keys of its own.

**SSIS custom setup: prefer express, reference Key Vault** -- the express form covers the common node preparation (environment variables, PowerShell version, licensed components, cmdkey credentials) without maintaining a script container, and its password and license fields each have a Key Vault reference alternative. When both inline and reference forms are set, Azure receives the INLINE value, silently winning. Every setup runs on every node start, so slow installers stretch every scale-out.

**The SSIS catalog outlives the runtime** -- `catalogInfo` creates the SSISDB database on YOUR Azure SQL server when the runtime first starts; deleting the runtime does not delete SSISDB. Pick the tier (or elastic pool -- mutually exclusive) here, but plan the database's backups, failover, and cost on the SQL server's side.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataFactory** | `dataFactoryId` | `status.outputs.data_factory_id` |
| **AzureSubnet** (SSIS express injection) | `azureSsis.expressVnetIntegration.subnetId` | `status.outputs.subnet_id` |
| **AzureVirtualNetwork** (SSIS standard injection) | `azureSsis.vnetIntegration.vnetId` | `status.outputs.virtual_network_id` |
| **AzureSubnet** (SSIS standard injection) | `azureSsis.vnetIntegration.subnetId` | `status.outputs.subnet_id` |
| **AzurePublicIp** (SSIS, exactly two) | `azureSsis.vnetIntegration.publicIps` | `status.outputs.public_ip_id` |
| **AzureDataFactoryLinkedService** (SSIS package stores, proxy staging, Key Vault references) | `azureSsis.packageStore[].linkedServiceName` and peers | `status.outputs.linked_service_name` |
| **AzureDataFactoryIntegrationRuntime** (SSIS proxy / shared registration) | `azureSsis.proxy.selfHostedIntegrationRuntimeName`, `selfHosted.rbacAuthorization.resourceId` | `status.outputs.integration_runtime_name` / `status.outputs.integration_runtime_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `integration_runtime_name` | The runtime's name inside the factory | Linked services pin connections to it; an SSIS runtime's proxy references a self-hosted one by this name |
| `integration_runtime_id` | The ARM ID (`{factory_id}/integrationRuntimes/{name}`) | The `rbacAuthorization.resourceId` of a linked self-hosted registration in another factory |
| `primary_authorization_key` | The join key for a PRIMARY self-hosted runtime (sensitive) | Handed to the agent installer on your machines |
| `secondary_authorization_key` | The rotation key for a PRIMARY self-hosted runtime (sensitive) | Agents migrate to it before the primary rotates |

The authorization keys are populated only for a primary self-hosted runtime -- Azure issues none for the managed flavors or for a linked registration.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Data-flow compute with a warm pool** -- the serverless Spark engine mapping data flows transform on, joined to the factory's managed virtual network with a small warm pool for back-to-back runs. Start from the **Data Flow Compute** preset.

**On-premises bridge** -- a free self-hosted registration; install the agent next to the data Azure cannot reach and hand it an authorization key. Share one primary runtime across factories with a linked (RBAC-authorized) registration instead of installing agents twice. Start from the **Self-Hosted Bridge** preset.

**SSIS lift-and-shift** -- the managed package runtime with an SSISDB catalog on your Azure SQL server; deploy existing SSIS projects unchanged, start the runtime for batch windows, and stop it after. Start from the **SSIS Runtime** preset.

**SSIS reaching on-premises data** -- pair the SSIS runtime's `proxy` block with a self-hosted runtime in the same factory, staging data through a storage linked service, instead of stretching the SSIS cluster into your network.

## Works With

- [**Azure Data Factory**](/infra-catalog/azure-data-factory) -- the factory the runtime lives in, referenced by `dataFactoryId`
- [**Azure Data Factory Linked Service**](/infra-catalog/azure-data-factory-linked-service) -- connections pin to a runtime by name; SSIS package stores, proxy staging, and Key Vault secret references all travel through linked services
- [**Azure Data Factory Data Flow**](/infra-catalog/azure-data-factory-data-flow) -- mapping data flows execute on the azure flavor
- [**Azure Data Factory Pipeline**](/infra-catalog/azure-data-factory-pipeline) -- activities run on the runtime their linked services resolve to
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- the injection target for SSIS virtual network integration
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the two static outbound addresses an injected SSIS runtime can present
