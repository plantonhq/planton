# Microsoft Fabric Capacity

Deploys a Microsoft Fabric capacity -- the billing and compute anchor of Microsoft Fabric: workspaces assign themselves to it, and its F-SKU sets how much compute every workload on it (lakehouses, warehouses, Power BI, real-time analytics) shares. The capacity is ARM's entire Fabric surface -- workspaces and the items inside them are managed through Microsoft's dedicated Fabric provider, portal, or APIs. A running capacity bills per hour from the moment it exists, and the SKU ladder spans three orders of magnitude, scaling up and down in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Fabric capacity** -- the capacity itself: its F-SKU (tier "Fabric", the only value Azure defines, sent by the platform) and its administrator list
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **A resource group** -- reference an AzureResourceGroup Infra Component or pass an existing group's name.
- **At least one administrator identity** -- an Entra user principal name (e.g. `admin@contoso.com`) or a service principal's object ID; Azure rejects a capacity created with none.
- **A supported subscription offer type** -- sponsorship and some trial/credit offers cannot create Fabric capacities; the rejection is an auth-shaped `401` that has nothing to do with your credentials (see Key Configuration).
- **A Fabric-supported region** -- Fabric is not available in every Azure region; check the availability list before choosing.

## Deploy

### Console

Open the deployment store, find **Microsoft Fabric Capacity**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the resource group, region, F-SKU, and administrators. Start from the **Starter Capacity** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFabricCapacity
metadata:
  name: analytics-capacity
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-analytics
      fieldPath: status.outputs.resource_group_name
  name: acmefabric
  region: eastus
  skuName: F2
  administrationMembers:
    - admin@acme-corp.com
```

```shell
planton apply -f fabric-capacity.yaml
```

This creates the smallest Fabric capacity (F2) with one administrator -- the hour meter starts when the deploy finishes. An Infra Job tracks the provisioning in real time.

### InfraChart

When the capacity and its resource group deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-analytics
      fieldPath: status.outputs.resource_group_name
  name: acmefabric
  region: eastus
  skuName: F64
  administrationMembers:
    - admin@acme-corp.com
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the capacity in it.

## Key Configuration

These are the most important decisions when configuring a Fabric capacity. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The meter runs from minute one** -- this is the rare catalog kind that costs real money at REST: the hour meter starts when the deploy finishes and stops only at delete, whether or not any workspace uses it. Pay-as-you-go F2 is roughly a third of a dollar per hour; F2048 scales that by a thousand. Treat capacity lifetime as an operational decision -- create development capacities when needed, delete them after, and never let a proof-of-concept capacity outlive its meeting.

**Size skuName with the dial, not with foresight** -- the F-SKU (`F2` through `F2048`, doubling per step) moves up AND down in place with no downtime and no replacement, so the right starting size is the smallest one that works (F2 for almost all development). Watch the Fabric Capacity Metrics app for throttling and smoothing-debt, and step the SKU when evidence -- not anticipation -- says so. The threshold that matters on the way up: F64 unlocks Copilot and lets Power BI free-license users consume shared content.

**administrationMembers are declared here, exercised there** -- the list is the bridge between the two worlds: identities declared on the ARM resource, exercised in the Fabric admin experience (assigning workspaces, managing settings). Keep it to a small platform group. The spec requires at least one entry at all times -- Azure's API technically allows clearing the list on update, but removing every administrator from a running paid capacity is a lockout, not a configuration.

**The capacity is ARM's whole Fabric story** -- everything inside Fabric (workspaces, lakehouses, warehouses, pipelines) is managed OUTSIDE this kind: Microsoft ships a dedicated Fabric Terraform provider, and the Fabric portal and APIs own the rest. This kind anchors billing and compute in your Azure estate; workspace governance is a Fabric-side discipline.

**Your subscription's offer type gates creation** -- unsupported offer types (sponsorships, some trial and credit offers) are rejected at create with `401 Unauthorized: "Unable to authorize with Azure Active Directory"` -- an auth-shaped error that has nothing to do with your credentials. The tell: the same error in every region, reproducible with your own Owner token against the ARM API directly, with the `Microsoft.Fabric` provider registered. The fix is account-level (a pay-as-you-go or enterprise offer), not a permissions change.

**The name is what workspace teams see** -- 3-63 lowercase letters and numbers, starting with a letter; Fabric workspaces select the capacity by this name, so make it recognizable. Name, region, and resource group are all ForceNew.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

`status.outputs` records identifiers only: `fabric_capacity_id` (the ARM ID) and `fabric_capacity_name`, which echoes the manifest's `name`. No catalog kind consumes them via ValueFromRef -- workspaces assign themselves to the capacity from the Fabric side (workspace settings, selecting the capacity by name), outside ARM's reach.

## Common Patterns

**Start at F2, grow on evidence** -- the smallest capacity for a first production workload or development environment, stepped up when the Capacity Metrics app shows throttling. Start from the **Starter Capacity** preset.

**Ephemeral development capacities** -- because the meter runs at rest, development and proof-of-concept capacities are created for the work and deleted after it; the in-place SKU dial makes recreating at a different size costless.

**One capacity per environment tier** -- a small capacity for development and a production capacity sized from measured usage, rather than one shared capacity where a runaway development workload can throttle production reports.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the capacity lives in

Everything downstream of the capacity -- workspaces, lakehouses, warehouses -- lives in Microsoft's own Fabric tooling rather than the catalog, so this kind composes with the rest of Fabric outside Planton.
