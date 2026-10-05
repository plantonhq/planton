# Azure Service Bus Namespace

Deploys an Azure Service Bus namespace -- the container and billing boundary for enterprise messaging. The namespace is where the pricing tier, network posture, encryption ownership, and authentication mode are set; the messaging entities themselves (queues, topics, subscriptions, scoped SAS rules, and the geo-DR pairing) are first-class catalog kinds that reference it, so application teams own their entities independently of the namespace's owner.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Bus Namespace** -- in the specified region and resource group, on the chosen tier (Basic, Standard, or Premium), with the public DNS identity `{namespace_name}.servicebus.windows.net`
- **Premium capacity** -- on the Premium tier only: the dedicated messaging units (1, 2, 4, 8, or 16) and the namespace partition layout (1, 2, or 4 -- fixed at creation)
- **Managed identity** -- when the `identity` block is configured: a system-assigned principal, user-assigned identity attachments, or both
- **Customer-managed-key encryption** -- when the `customerManagedKey` block is configured (Premium only): messaging data encrypts under your Key Vault key, unwrapped by the attached user-assigned identity
- **Namespace firewall** -- when the `networkRuleSet` block is configured (Premium only): a default action plus admitted IP ranges and VNet service-endpoint subnets
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

The namespace deliberately contains no messaging entities at creation: queues, topics, subscriptions, scoped SAS rules, and the geo-DR pairing are their own kinds, each referencing this namespace's `namespace_id` output.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the namespace will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A globally unique namespace name** -- 6-50 characters of letters, numbers, and hyphens, starting with a letter and ending with a letter or number. It becomes the endpoint `{name}.servicebus.windows.net`, and Azure reserves the suffixes `-sb` and `-mgmt`.
- **For customer-managed keys (Premium)** -- an AzureKeyVault with purge protection, an AzureKeyVaultKey in it, and an AzureUserAssignedIdentity holding wrap/unwrap on the vault.

## Deploy

### Console

Open the deployment store, find **Azure Service Bus Namespace**, and click **Deploy**. The creation wizard walks you through placement, the pricing tier (with the Premium capacity pair), managed identity, encryption ownership (Premium), the SAS-vs-keyless authentication posture, network access, and governance tags. Start from the **Standard Namespace** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusNamespace
metadata:
  name: order-bus
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  namespaceName: acme-order-bus
  sku: STANDARD
  tags:
    purpose: order-processing
```

```shell
planton apply -f servicebus-namespace.yaml
```

This creates a Standard-tier namespace (leaving `sku` out entirely is also valid -- Azure deploys STANDARD when the spec records no tier); queues and topics arrive as their own kinds afterward, referencing this namespace. An Infra Job tracks the provisioning in real time.

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

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the namespace with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Service Bus namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier** -- STANDARD (Azure's default when unspecified) is the full-featured multi-tenant tier: topics with subscriptions, sessions, duplicate detection, transactions. BASIC supports queues only. PREMIUM runs dedicated messaging units and is the only tier with VNet integration, customer-managed keys, geo-DR, partitioning, and 100 MB messages. BASIC <-> STANDARD updates in place; moving into or out of PREMIUM **replaces the namespace and every entity in it**.

**The Premium capacity pair** -- PREMIUM requires both `capacity` (messaging units: 1, 2, 4, 8, or 16 -- scaling updates in place) and `premiumMessagingPartitions` (1, 2, or 4). The partition layout is **fixed at creation**, and every queue and topic in a partitioned namespace must enable partitioning itself.

**Authentication posture** -- every namespace carries a root SAS rule whose keys surface as sensitive outputs. Set `localAuthEnabled: false` for the keyless posture: every SAS key (including the root keys) becomes inert, and clients authenticate with Microsoft Entra identities holding data-plane roles granted via AzureRoleAssignment.

**Network posture** -- `publicNetworkAccessEnabled: false` removes the public endpoint entirely (pair it with an AzurePrivateEndpoint targeting subresource `namespace`). The Premium-only `networkRuleSet` firewall scopes data-plane access by IP range and subnet; Azure rejects DENY with no admitted source.

**Customer-managed keys (Premium)** -- the `customerManagedKey` block encrypts messaging data under your Key Vault key. It requires a user-assigned identity attached via the `identity` block, and **once enabled it cannot be removed** -- dropping the block replaces the namespace.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds[]` | `status.outputs.identity_id` |
| **AzureKeyVaultKey** | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |
| **AzureUserAssignedIdentity** | `customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| **AzureSubnet** | `networkRuleSet.networkRules[].subnetId` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace_id` | Azure Resource Manager ID of the namespace | The parent reference for every Service Bus entity kind, the scope for namespace-wide data-plane role assignments, and the AzurePrivateEndpoint target (subresource: `namespace`) |
| `namespace_name` | Name of the Service Bus namespace | SDK configuration, monitoring dashboards |
| `endpoint` | Service Bus endpoint URL (`https://{name}.servicebus.windows.net:443/`) | Service Bus SDK connection configuration |
| `identity_principal_id` | The system-assigned identity's principal ID (when configured) | AzureRoleAssignment grants for the namespace's own access |
| `default_primary_connection_string` | The root SAS rule's primary connection string (sensitive -- full manage rights) | Quick starts and break-glass use only |
| `default_secondary_connection_string` | The root SAS rule's secondary connection string (sensitive) | The rotation partner for the primary |
| `default_primary_key` | The root SAS rule's primary key (sensitive) | SDKs taking key and key name separately |
| `default_secondary_key` | The root SAS rule's secondary key (sensitive) | The rotation partner for the primary key |

Production workloads should mint least-privilege credentials with AzureServiceBusAuthorizationRule -- or go keyless -- rather than distributing the root credentials.

## Common Patterns

**Standard namespace as the default** -- the full-featured multi-tenant tier (topics, sessions, duplicate detection) fits most production workloads at multi-tenant pricing; entities arrive later as their own kinds. The trade: no VNet integration, CMK, or geo-DR -- and moving to PREMIUM later replaces the namespace and every entity in it, so pick the tier deliberately. Start from the **Standard Namespace** preset.

**Premium isolated for latency-sensitive workloads** -- dedicated messaging units with the namespace firewall on DENY and admitted subnets only, for predictable latency and network isolation. Remember the partition layout is fixed at creation, and DENY requires at least one admitted source or Azure refuses the configuration. Start from the **Premium Isolated Namespace** preset.

**Keyless with Entra identities** -- `localAuthEnabled: false` makes every SAS key inert (including the root keys in this kind's outputs), so no static credential exists to leak or rotate; clients hold Entra data-plane roles (Azure Service Bus Data Owner/Sender/Receiver) granted via AzureRoleAssignment. The trade: every client must support Entra authentication -- legacy connection-string consumers break. Start from the **Keyless Namespace (Entra-Only)** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the namespace is created
- [**Azure Service Bus Queue**](/infra-catalog/azure-service-bus-queue) -- point-to-point messaging entities living in this namespace
- [**Azure Service Bus Topic**](/infra-catalog/azure-service-bus-topic) -- publish-subscribe distribution, with [**Azure Service Bus Subscription**](/infra-catalog/azure-service-bus-subscription) under a topic
- [**Azure Service Bus Authorization Rule**](/infra-catalog/azure-service-bus-authorization-rule) -- least-privilege SAS credentials scoped to the namespace or a single entity
- [**Azure Service Bus Disaster Recovery Config**](/infra-catalog/azure-service-bus-disaster-recovery-config) -- the geo-DR alias pairing two Premium namespaces
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed key Premium namespaces encrypt messaging data under
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- attached via the identity block; unwraps customer-managed keys
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- admitted to the Premium firewall via service endpoints
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- takes the namespace off the public internet (subresource: `namespace`)
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants Entra identities data-plane roles for the keyless posture
