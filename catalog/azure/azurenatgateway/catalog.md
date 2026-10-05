# Azure NAT Gateway

Deploys an Azure NAT Gateway -- the managed source-network-address-translation (SNAT) service that gives every workload in its attached subnets stable, scalable outbound internet connectivity, and the production answer to Azure retiring implicit default outbound access. The gateway is deliberately just the gateway: the public addresses it SNATs through are first-class AzurePublicIp / AzurePublicIpPrefix resources referenced by ID, and the subnets it serves declare the attachment themselves (each AzureSubnet's `natGatewayId`), matching Azure's one-gateway-many-subnets model. A gateway with no addresses deploys but cannot translate anything — associate at least one IP or prefix for it to carry traffic.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **NAT Gateway** -- a Standard or StandardV2 SKU gateway in the specified region and resource group, with configurable idle timeout and (for Standard) optional availability-zone pinning
- **Public IP Associations** -- links to the referenced `publicIpIds`, each adding 64,512 SNAT ports
- **Public IP Prefix Associations** -- links to the referenced `publicIpPrefixIds`, contiguous reserved ranges for scalable, allowlistable egress
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the gateway

Subnet attachments are NOT created here -- each AzureSubnet declares its own `natGatewayId`.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the NAT Gateway will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Public addresses** -- a gateway with no addresses deploys but cannot translate anything. Reference AzurePublicIp / AzurePublicIpPrefix Infra Components (a StandardV2 gateway needs StandardV2 addresses).
- **Region alignment** -- the gateway only serves subnets in its own region.

## Deploy

### Console

Open the deployment store, find **Azure NAT Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard NAT Gateway** preset in the [Presets](#presets) tab to pre-populate a zonal gateway SNATing through one referenced public IP.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureNatGateway
metadata:
  name: prod-nat
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: prod-nat
  zones:
    - "1"
  publicIpIds:
    - valueFrom:
        kind: AzurePublicIp
        name: nat-egress-ip
        fieldPath: status.outputs.public_ip_id
  idleTimeoutInMinutes: 10
```

```shell
planton apply -f azure-nat-gateway.yaml
```

This creates a zonal Standard gateway SNATing through one public IP with a 10-minute idle timeout. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the gateway to its resource group and addresses -- and wire each subnet to the gateway:

```yaml
# On the AzureNatGateway:
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  publicIpIds:
    - valueFrom:
        kind: AzurePublicIp
        name: nat-egress-ip
        fieldPath: status.outputs.public_ip_id

# On each AzureSubnet that should egress through it:
spec:
  natGatewayId:
    valueFrom:
      kind: AzureNatGateway
      name: prod-nat
      fieldPath: status.outputs.nat_gateway_id
```

The InfraPipeline resolves the dependency graph, deploys the group and addresses first, then the gateway, then the subnets that attach it.

## Key Configuration

These are the most important decisions when configuring a NAT gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU** -- `skuName` unset applies Azure's default (STANDARD): a zonal gateway, optionally pinned to one availability zone via `zones`. STANDARD_V2 is zone-redundant automatically (`zones` must be empty -- the spec enforces it) and requires StandardV2 public IPs/prefixes. Fixed at creation.

**Public addresses** -- `publicIpIds` references individual addresses (64,512 SNAT ports each); `publicIpPrefixIds` references contiguous reserved ranges -- the scalable, allowlistable alternative (a /28 prefix multiplies capacity by 16 in one block). Both updatable in place.

**Idle timeout** -- `idleTimeoutInMinutes` controls how long an idle outbound TCP connection's SNAT port stays reserved (4-120, Azure default 4). Higher values hold ports longer and hasten SNAT exhaustion -- raise only for long-lived idle connections that must not re-establish.

**Zone resilience** -- a STANDARD gateway lives in one zone (or non-zonal when `zones` is empty); zone-resilient architectures deploy one gateway per zone with per-zone subnets, or use STANDARD_V2 for built-in redundancy.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzurePublicIp** (repeated) | `publicIpIds` | `status.outputs.public_ip_id` |
| **AzurePublicIpPrefix** (repeated) | `publicIpPrefixIds` | `status.outputs.public_ip_prefix_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `nat_gateway_id` | Azure Resource Manager ID of the gateway | AzureSubnet's `natGatewayId` references it to attach the gateway to a subnet |
| `resource_guid` | The immutable GUID ARM assigns | Correlating with Azure billing, monitoring, or support data that keys on the GUID |

`status.outputs` also echoes `nat_gateway_name` back, but attachments travel by ARM ID — nothing downstream consumes the gateway by name.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard zonal gateway** -- A zonal Standard gateway SNATing through one referenced public IP: stable outbound with a known source address for every attached subnet. Start from the **Standard NAT Gateway** preset.

**Prefix-backed SNAT range** -- A gateway SNATing through a public IP prefix: one contiguous, allowlistable range that scales SNAT capacity for high-throughput estates. Start from the **NAT Gateway with a Prefix SNAT Range** preset.

**Zone-redundant V2** -- Azure's next-generation StandardV2 gateway: zone-redundant automatically, no zone pinning, StandardV2 addresses. Start from the **Zone-Redundant StandardV2 NAT Gateway** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the gateway is created in
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the addresses the gateway SNATs through
- [**Azure Public IP Prefix**](/infra-catalog/azure-public-ip-prefix) -- contiguous reserved ranges for scalable egress
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- attaches the gateway via its `natGatewayId` to route the subnet's egress
