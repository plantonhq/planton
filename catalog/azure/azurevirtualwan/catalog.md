# Azure Virtual WAN

Deploys a Virtual WAN -- the free, lightweight umbrella object of Azure's managed hub-and-spoke networking. Regional virtual hubs (and the VPN/ExpressRoute gateways on them) are separate resources that reference this WAN; the WAN itself carries the global transit policy.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual WAN** -- the ARM policy object with its type (Standard/Basic), VPN-encryption and branch-to-branch settings, and Office 365 breakout category
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the WAN

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the WAN will be created.

## Deploy

### Console

Open the deployment store, find **Azure Virtual WAN**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard WAN** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualWan
metadata:
  name: global-wan
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: global-wan
```

```shell
planton apply -f azure-virtual-wan.yaml
```

This creates a Standard-tier WAN with ARM's defaults -- branch-to-branch transit on, no Office 365 breakout. The WAN provisions in minutes and is free by itself; hubs and gateways carry the cost. An Infra Job tracks the provisioning in real time.

### InfraChart

In a hub-and-spoke chart, the WAN is the root: WAN → virtual hub(s) → hub connections and gateways, each wiring to the previous by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: global-wan
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then the WAN, then the hubs that reference its `virtual_wan_id` output.

## Key Configuration

These are the most important decisions when configuring a WAN. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Type** -- "Standard" (the default) is the full mesh: ExpressRoute, site-to-site and point-to-site VPN, hub-to-hub transit. "Basic" is a constrained legacy tier (site-to-site only, Basic hubs) that can be UPGRADED to Standard in place but never downgraded.

**Branch-to-branch traffic** -- on by default (ARM's default); set `allowBranchToBranchTraffic: false` only when branches must not reach each other through the WAN.

**Office 365 breakout** -- `office365LocalBreakoutCategory` declares which O365 traffic exits at local branch internet breakouts instead of transiting the WAN: NONE (default), OPTIMIZE, OPTIMIZE_AND_ALLOW, or ALL.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `virtual_wan_id` | Azure Resource Manager ID of the WAN | A virtual hub's `virtualWanId` |

The only other output, `virtual_wan_name`, echoes the WAN's name back; no downstream Infra Component consumes it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard WAN** -- the full-mesh default. Start from the **Standard WAN** preset.

**Isolated branches** -- branch-to-branch transit off for hub-and-spoke-only reachability. Start from the **Isolated Branches WAN** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the WAN is created in
- [**Azure Virtual Hub**](/infra-catalog/azure-virtual-hub) -- the regional routers that reference this WAN's `virtual_wan_id`
