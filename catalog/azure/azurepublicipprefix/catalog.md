# Azure Public IP Prefix

Deploys an Azure Public IP Prefix — a reserved, CONTIGUOUS range of public IP addresses. Individual public IPs allocated from a prefix are guaranteed to come from the same known range, which is what lets partners and firewalls allowlist a single CIDR instead of chasing individual addresses, and what gives a NAT gateway a predictable, scalable SNAT range. The prefix is essentially immutable: everything except tags is fixed at creation, a replacement is a **different range**, and a prefix cannot be deleted while any of its addresses are in use.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Public IP Prefix** -- the reserved range, sized by your CIDR length (Azure's default is /28 — 16 addresses), with the SKU, tier, IP version, and zone anchoring you chose
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The actual CIDR is assigned by Azure at creation and surfaces as the `ip_prefix` output — the value partners allowlist. Azure bills every reserved address from creation, allocated or not.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the prefix will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A size plan**: running out later means a SECOND prefix (and a second allowlist entry) — size for the growth you expect. IPv4 spans /21 (2,048 addresses) to /31 (2).

## Deploy

### Console

Open the deployment store, find **Azure Public IP Prefix**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **NAT Gateway SNAT Range** preset in the [Presets](#presets) tab for the flagship zone-redundant egress range.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePublicIpPrefix
metadata:
  name: prod-egress
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-network-hub"
  name: prod-egress
  prefixLength: 28
  zones:
    - "1"
    - "2"
    - "3"
  tags:
    cost-center: platform-network
```

```shell
planton apply -f prefix.yaml
```

This reserves 16 contiguous zone-redundant addresses; the assigned CIDR lands in `status.outputs.ip_prefix` — hand that one value to every partner allowlist. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the prefix to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-hub
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then reserves the range — and the NAT gateway and public IPs that consume it reference this prefix's `public_ip_prefix_id`.

## Key Configuration

These are the most important decisions when configuring a Public IP Prefix. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Prefix length** -- the CIDR length decides how many contiguous addresses Azure reserves: /28 (16, the default), /26 (64), /21 (2,048 — the IPv4 maximum range). Smaller numbers reserve bigger ranges, and every reserved address bills from creation. Fixed at creation — growth means a second prefix.

**SKU and tier** -- unspecified applies Azure's defaults (Standard, Regional) — correct for virtually everything. `STANDARD_V2` exists for StandardV2 NAT gateway association; `GLOBAL` exists solely for cross-region load balancer frontends and **must keep the STANDARD SKU** (ARM rejects StandardV2 with Global — the wizard holds this rule live).

**Zones** -- all three zones make the range zone-redundant (the production default); a single zone pins it; empty leaves the choice to Azure. Fixed at creation.

**Custom IP Prefix (BYOIP)** -- leave empty to allocate from Microsoft's pool (the overwhelmingly common case). Set the ARM ID of an onboarded Custom IP Prefix only when your org brought its own IP space to Azure.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `public_ip_prefix_id` | Azure Resource Manager ID of the prefix | AzurePublicIp `publicIpPrefixId` (allocate from the range), AzureNatGateway `publicIpPrefixIds` (SNAT association) |
| `ip_prefix` | The actual reserved CIDR, e.g. "20.42.0.16/28" | Partner and firewall allowlists — the range's whole reason to exist |

The outputs also carry `public_ip_prefix_name`, the prefix resource's name -- consumers reference the prefix by ARM ID, so the name has no ValueFromRef consumer.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**NAT SNAT range** -- a zone-redundant /28 associated with a NAT gateway: predictable egress addresses that scale SNAT ports with the fleet. Start from the **NAT Gateway SNAT Range** preset.

**Partner allowlist** -- a dedicated range for one integration, so the partner allowlists one CIDR and revoking the integration later touches nothing else. Start from the **Partner Firewall Allowlist** preset.

**From-prefix public IP** -- reserve the range here, then allocate individual AzurePublicIp resources from it — each guaranteed to come from the allowlisted CIDR. Start from the **Public IP From Prefix** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the prefix is created
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- allocates individual addresses from the range by referencing its `public_ip_prefix_id`
- [**Azure NAT Gateway**](/infra-catalog/azure-nat-gateway) -- associates the whole prefix for outbound SNAT, the flagship consumption
- [**Azure Load Balancer**](/infra-catalog/azure-load-balancer) -- fronts traffic with public IPs drawn from the allowlisted range
