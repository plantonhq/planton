# Azure Availability Set

Deploys an availability set -- the classic placement grouping that spreads VMs across separate fault and update domains so one hardware failure or maintenance window cannot take them all down. The set is free, and its entire configuration is fixed at creation: only tags update in place, and VMs join the set only when they themselves are created. In regions with availability zones, prefer zones -- a VM uses a zone OR an availability set, never both; the set remains the right tool in regions without zones and for classic lift-and-shift topologies.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Availability set** -- the placement grouping with its fault-domain count (independent power/network/rack groups, default 3), update-domain count (planned-maintenance batches, default 5), managed-disk alignment (default on), and optional proximity placement group association
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource kind, resource ID); a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** in the target region -- referenced through `resourceGroup` as a literal name or an AzureResourceGroup ValueFromRef. VMs joining the set must live in the same resource group and region.
- **Region fault-domain capacity** (only when setting `platformFaultDomainCount`) -- some regions support fewer than 3 fault domains, and Azure rejects a count the region cannot provide at create time.

## Deploy

### Console

Open the deployment store, find **Azure Availability Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Classic Web Tier** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAvailabilitySet
metadata:
  name: web-avset
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  name: web-avset
  region: eastus
```

```shell
planton apply -f availability-set.yaml
```

This creates an availability set with the provider defaults -- 5 update domains, 3 fault domains, managed-disk alignment on -- ready for VMs to join at their creation. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the resource group:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  name: web-avset
  region: eastus
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the set with the resolved group name.

## Key Configuration

These are the most important decisions when configuring an availability set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zones or a set -- decide before the first VM** -- An availability zone survives a datacenter failure; an availability set survives a rack, power, or switch failure inside one datacenter. A VM is placed in a zone or a set, never both, so this is an architecture decision made per tier. In zoned regions, new designs should pin VMs to zones and skip the set; the set earns its place in regions without zones and in classic topologies lifted as-is.

**Everything except tags is a one-way door** -- Every spec field carries ForceNew semantics: changing the resource group, name, region, domain counts, managed alignment, or proximity placement group destroys and recreates the set -- and since VMs join only at their own creation, rebuilding a set means recreating every VM in it. Treat the set like a subnet, not like a tag: decide its shape before the first VM exists.

**Fault and update domain counts** -- Leaving `platformFaultDomainCount` and `platformUpdateDomainCount` unset takes the provider defaults (3 and 5), the right shape for almost every tier. Lower the fault-domain count only when the target region cannot provide 3; raise update domains only when a tier has enough VMs that a fifth of them rebooting at once during planned maintenance is too many.

**Managed alignment stays on** -- `managed` defaults to true, aligning fault domains with the VMs' managed-disk storage so a storage-cluster failure does not cross your compute fault domains. The false setting exists for the unmanaged-disk era; every managed-disk VM belongs in a managed set. Leave the field unset.

**Proximity placement is a latency/resilience trade** -- `proximityPlacementGroupId` (a plain ARM ID; proximity placement groups are not modeled as a Planton kind) co-locates the set's VMs for minimal inter-VM latency, at the cost of shrinking the hardware pool the set can spread across. Set it only when a latency-sensitive tier has measured the need.

**Two VMs minimum, or the set is theater** -- Azure's classic 99.95% multi-VM SLA starts at two VMs in the set; a single-VM availability set provides nothing except a future constraint. If a tier has two or more VMs, put them all in -- a tier half-in, half-out fails together through the half that shares hardware.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `availability_set_id` | Azure Resource Manager ID of the set | AzureVirtualMachine's `availability.availabilitySetId` -- how VMs join the set at their creation |

The only other output, `availability_set_name`, echoes the configured name back for reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Classic multi-VM web tier** -- one availability set per load-balanced tier, provider-default domains, two or more VMs joining at creation. This is the pattern that carries the classic multi-VM SLA. Start from the **Classic Web Tier** preset.

**Latency-pinned tier** -- the same shape with `proximityPlacementGroupId` set, co-locating the tier's VMs for minimal inter-VM latency. Accept that the placement group narrows the hardware pool and can make allocations fail in constrained regions.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the set lives in; VMs joining the set must live in the same group and region
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- joins the set at creation through `availability.availabilitySetId`, which defaults to this component's `availability_set_id` output
