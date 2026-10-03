# Azure Event Hub Namespace

Deploys an Azure Event Hubs namespace -- the container and billing boundary for high-throughput event streaming. The namespace carries the pricing tier, throughput capacity, network posture, and authentication mode; the streaming entities (event hubs, consumer groups, SAS rules, schema groups, geo-DR pairings) are first-class Infra Components that reference it. The namespace name becomes the endpoint `{name}.servicebus.windows.net` and the Kafka bootstrap host on port 9093, so existing Kafka producers and consumers connect without code changes (the Basic tier has no Kafka endpoint).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Hubs Namespace** -- a namespace in the specified region and resource group with the chosen SKU tier (Basic, Standard, or Premium), throughput or processing-unit capacity, and optional auto-inflate elastic scaling
- **Managed Identity** -- created only when the `identity` block is configured; a system-assigned and/or user-assigned Entra identity the namespace authenticates with (capture to Storage, customer-managed keys)
- **Namespace Firewall** -- created only when the `networkRuleSets` block is configured (Standard/Premium); a default action plus admitted IP ranges, VNet service-endpoint subnets, and the trusted-Microsoft-services bypass
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance, merged with any user tags

The event streams themselves are separate Infra Components: deploy an **AzureEventHub** per stream against this namespace's outputs, an **AzureEventHubConsumerGroup** per independent reader, and an **AzureEventHubAuthorizationRule** per scoped credential.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Event Hubs namespace will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A globally unique namespace name** -- the name becomes the endpoint `{name}.servicebus.windows.net` (and the Kafka bootstrap host on port 9093) and must be unique across all of Azure.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Namespace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Streaming Namespace** preset in the [Presets](#presets) tab for the common production entry point.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubNamespace
metadata:
  name: platform-events
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  namespaceName: acme-platform-events
  sku: STANDARD
  capacity: 2
  autoInflateEnabled: true
  maximumThroughputUnits: 10
```

```shell
planton apply -f eventhub-namespace.yaml
```

This creates a Standard-tier namespace starting at 2 throughput units with auto-inflate growing it to at most 10 under load. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the namespace to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the Event Hub namespace with the resolved values -- and the namespace's own outputs feed the AzureEventHub streams deployed after it.

## Key Configuration

These are the most important decisions when configuring an Event Hub namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier** -- STANDARD (the default when unset) supports the Kafka endpoint, up to 20 consumer groups per hub, 7-day retention, and auto-inflate. BASIC limits to a single consumer group, 1-day retention, no Kafka, and no firewall. PREMIUM provides reserved processing units with predictable latency and extended retention. BASIC <-> STANDARD updates in place; moving into or out of PREMIUM **replaces the namespace and every entity in it**.

**Capacity and auto-inflate** -- on BASIC/STANDARD, `capacity` is throughput units (1-40; each TU provides 1 MB/s ingress and 2 MB/s egress). Enable `autoInflateEnabled` with `maximumThroughputUnits` to grow automatically under load -- auto-inflate only scales UP; trim `capacity` manually after a spike. On PREMIUM, `capacity` is processing units (Azure sells 1, 2, 4, 8, or 16).

**Authentication posture** -- every namespace carries a root SAS rule whose keys surface as sensitive outputs. Set `localAuthenticationEnabled: false` for the keyless posture: all SAS keys stop working and clients authenticate with Microsoft Entra identities holding data-plane roles (granted via AzureRoleAssignment against the namespace ID).

**Network access** -- set `publicNetworkAccessEnabled: false` to restrict the namespace to private endpoints (AzurePrivateEndpoint, subresource `namespace`) or admitted VNet service endpoints. On Standard/Premium, add a `networkRuleSets` block for the data-plane firewall; a DENY default action requires at least one admitted IP rule or subnet.

**Dedicated cluster** -- reference an AzureEventHubCluster via `dedicatedClusterId` for single-tenant capacity, up to 1024 partitions per hub, and 90-day retention. A namespace can never move on or off a cluster in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureEventHubCluster** | `dedicatedClusterId` | `status.outputs.cluster_id` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| **AzureSubnet** | `networkRuleSets.virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace_id` | Azure Resource Manager ID of the namespace | The parent reference for every Event Hubs satellite kind, namespace-scoped RBAC grants, private endpoints, diagnostic settings |
| `namespace_name` | Name of the Event Hubs namespace | SDK configuration, Kafka bootstrap host derivation, monitoring dashboards |
| `identity_principal_id` | The system-assigned identity's principal ID (empty unless the identity includes SYSTEM_ASSIGNED) | AzureRoleAssignment grants for capture (Storage) and CMK (Key Vault) |
| `default_primary_connection_string` | The root SAS rule's primary connection string (sensitive; full manage rights) | Quick starts and break-glass access -- prefer AzureEventHubAuthorizationRule credentials in production |
| `default_secondary_connection_string` | The root rule's secondary connection string (sensitive) | The rotation partner |
| `default_primary_key` | The root rule's primary key (sensitive) | SDKs taking key and key name separately |
| `default_secondary_key` | The root rule's secondary key (sensitive) | The rotation partner |
| `default_primary_connection_string_alias` | The primary connection string addressing the geo-DR alias hostname (sensitive; empty without a geo-DR pairing) | Clients that must survive a regional failover |
| `default_secondary_connection_string_alias` | The secondary alias connection string (sensitive) | The rotation partner |

With `localAuthenticationEnabled: false`, the six credential outputs still populate but are **inert** -- Azure rejects SAS authentication namespace-wide.

## Common Patterns

**Standard streaming** -- an explicit STANDARD namespace with elastic throughput (auto-inflate to a ceiling): the default starting point for telemetry, logging, and Kafka migrations. Remember auto-inflate only scales UP -- trim `capacity` manually after a traffic spike or the higher bill persists. Start from the **Standard Streaming Namespace** preset.

**Locked-down keyless** -- SAS authentication disabled (Entra-only data plane) plus a DENY firewall admitting only named sources, with trusted Microsoft services still delivering. No connection strings to rotate, leak, or vault; the trade is that every client needs an Entra identity and a data-plane role. Start from the **Locked-Down Keyless Namespace** preset.

**Premium isolated** -- reserved processing units with a system-assigned identity, for latency-sensitive streams that must not share throughput with other tenants: the isolation tier below a whole dedicated cluster. Remember moving into or out of PREMIUM replaces the namespace and every entity in it. Start from the **Premium Isolated Namespace** preset.

## Works With

The namespace is the hub of a satellite family -- each streaming entity deploys as its own Infra Component referencing `status.outputs.namespace_id`:

- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- the event stream: partition layout, retention, capture-to-storage
- [**Azure Event Hub Consumer Group**](/infra-catalog/azure-event-hub-consumer-group) -- independent read cursors per downstream system
- [**Azure Event Hub Authorization Rule**](/infra-catalog/azure-event-hub-authorization-rule) -- least-privilege SAS credentials scoped to the namespace or one hub
- [**Azure Event Hub Schema Group**](/infra-catalog/azure-event-hub-schema-group) -- the schema registry
- [**Azure Event Hub Disaster Recovery Config**](/infra-catalog/azure-event-hub-disaster-recovery-config) -- the geo-DR alias pairing two namespaces
- [**Azure Event Hub Cluster**](/infra-catalog/azure-event-hub-cluster) -- dedicated single-tenant hardware this namespace can be placed on via `dedicatedClusterId`
- [**Azure Event Hub Namespace Customer Managed Key**](/infra-catalog/azure-event-hub-namespace-customer-managed-key) -- BYOK encryption for namespaces on a dedicated cluster
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the namespace is created
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- pre-created identities for capture and customer-managed-key compositions
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- subnets admitted through the firewall via the Microsoft.EventHub service endpoint
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- VNet-private access when the public endpoint is disabled (subresource `namespace`)
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- Entra data-plane roles (Data Owner/Sender/Receiver) for the keyless posture
