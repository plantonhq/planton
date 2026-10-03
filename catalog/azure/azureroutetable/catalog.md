# Azure Route Table

Deploys an Azure Route Table — a reusable set of user-defined routes (UDRs) that overrides Azure's default system routing for every subnet attached to it. The building block of forced tunneling, firewall egress, network-virtual-appliance steering, and black-hole guardrails. The attachment is inverted by Azure's own model: a **subnet declares which table it uses** (an AzureSubnet's `routeTableId` references this table's `route_table_id` output), so one table serves many subnets and editing its routes re-routes all of them at once.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Route Table** -- the named table with BGP route propagation set per your dial (Azure defaults to enabled)
- **User-Defined Routes** -- each route managed inline as part of the table (a route has no life of its own in Azure); the table is the authoritative route list, so a deploy removes routes added out-of-band
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The subnet-side attachment is NOT created here — which subnets adopt this table lives on each referenced AzureSubnet.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the table will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **For virtual-appliance routes**: the appliance's private IP — reference an AzureFirewall's `private_ip_address` output (the hub-spoke seam), or pass the literal IP of a non-firewall NVA.

## Deploy

### Console

Open the deployment store, find **Azure Route Table**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Firewall Egress** preset in the [Presets](#presets) tab for the flagship 0.0.0.0/0-via-firewall shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRouteTable
metadata:
  name: egress-via-firewall
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-network-hub"
  name: egress-via-firewall
  routes:
    - name: default-via-firewall
      addressPrefix: 0.0.0.0/0
      nextHopType: VIRTUAL_APPLIANCE
      nextHopInIpAddress:
        value: "10.0.1.4"
  bgpRoutePropagationEnabled: false
```

```shell
planton apply -f route-table.yaml
```

This creates a route table that sends every attached subnet's internet-bound traffic through the firewall at 10.0.1.4, with BGP propagation disabled so learned on-premises routes cannot bypass it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the table to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-hub
      fieldPath: status.outputs.resource_group_name
  routes:
    - name: default-via-firewall
      addressPrefix: 0.0.0.0/0
      nextHopType: VIRTUAL_APPLIANCE
      nextHopInIpAddress:
        valueFrom:
          kind: AzureFirewall
          name: hub-firewall
          fieldPath: status.outputs.private_ip_address
```

The InfraPipeline resolves the dependency graph, deploys the resource group and firewall first, then provisions the table with the firewall's real address — and the subnets that adopt it reference this table's `route_table_id`.

## Key Configuration

These are the most important decisions when configuring a Route Table. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Routes** -- each route steers traffic bound for its `addressPrefix` (a CIDR block, or an Azure service tag like `"ApiManagement"`) to its `nextHopType`: `VIRTUAL_APPLIANCE` (with `nextHopInIpAddress` — required exactly there, rejected everywhere else), `VIRTUAL_NETWORK_GATEWAY` (on-premises over VPN/ExpressRoute), `INTERNET`, `VNET_LOCAL`, or `NONE` (black hole). The most specific prefix wins; among equals a user-defined route beats a system route. An **empty routes list is valid and common** — attach the empty table first, add routes as the topology grows.

**BGP propagation** -- `bgpRoutePropagationEnabled` governs whether routes learned from on-premises via BGP reach attached subnets. Azure defaults to true; **disable it on forced-tunnel and firewall-egress tables** so a learned route cannot silently bypass your appliance. Leave it unset to record no opinion (Azure's default applies).

**Name** -- the routing policy's identity ("egress-via-firewall"), unique within the resource group. Renaming replaces the table, detaching it from every subnet until the replacement re-attaches.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureFirewall** (optional, per appliance route) | `routes[].nextHopInIpAddress` | `status.outputs.private_ip_address` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `route_table_id` | Azure Resource Manager ID of the table | AzureSubnet `routeTableId` — the attachment seam that adopts this table's routing |

The `route_table_name` output echoes the spec's name back for convenience; it has no downstream wiring story of its own.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Firewall egress** -- 0.0.0.0/0 forwarded to the hub firewall's private IP by reference, propagation disabled — the hub-and-spoke standard. Start from the **Firewall Egress** preset.

**Forced tunnel** -- 0.0.0.0/0 to the virtual network gateway, sending all egress on-premises for compliance regimes that require inspection there. Start from the **Forced Tunneling On-Premises** preset.

**Black-hole guardrails** -- drop-routes for prefixes a workload must never reach — a routing-layer guardrail that survives NSG mistakes. Start from the **Black-Hole Guardrails** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the table is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- adopts this table by referencing its `route_table_id` output (the attachment lives on the subnet, one table serving many)
- [**Azure Firewall**](/infra-catalog/azure-firewall) -- the classic virtual-appliance next hop, referenced by its `private_ip_address` output
- [**Azure NAT Gateway**](/infra-catalog/azure-nat-gateway) -- the complementary egress path: NAT for outbound internet, this table for steering and inspection
