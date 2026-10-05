# Azure IP Group

Deploys an Azure IP Group — a named SET of IP addresses and CIDR ranges that Azure Firewall rules reference by name instead of repeating address lists rule by rule. Manage the set once, reference it from many rules across many firewall policies: adding or removing an address immediately retargets every rule that references the group, so the address plan changes in one place and the policy never moves. Like all of Planton's composition anchors, the referencing is inverted — the group holds only addresses; firewall policy rules point at it (`source_ip_groups` / `destination_ip_groups`).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **IP Group** -- the named address set with its entries (single IPs and CIDR blocks, up to 5,000 per group)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The rules that reference the group are NOT created here — they live on firewall policies and rule collection groups, each with its own lifecycle. An **empty group is legal**: a placeholder anchor rules can reference before the address plan is final.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the group will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **An address-set plan by INTENT**: one group per policy meaning ("branch-offices", "on-prem-datacenter", "blocked-scanners") — the names become the vocabulary your firewall rules read as.

## Deploy

### Console

Open the deployment store, find **Azure IP Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Branch-Office Address Set** preset in the [Presets](#presets) tab for the classic trusted-ranges set.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureIpGroup
metadata:
  name: branch-offices
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-network-hub"
  name: branch-offices
  cidrs:
    - "198.51.100.0/24"
    - "203.0.113.0/24"
  tags:
    cost-center: network-security
```

```shell
planton apply -f ip-group.yaml
```

This creates the `branch-offices` group with two trusted ranges -- every firewall rule referencing it matches them, and a new branch joins by adding one address here, with no policy edit. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the group to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-hub
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the group — and the firewall policy rules that use it reference this group's `ip_group_id`.

## Key Configuration

These are the most important decisions when configuring an IP Group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Addresses** -- `cidrs` accepts single addresses ("203.0.113.7") and CIDR blocks ("10.10.0.0/16"), up to 5,000 entries per group. Entries update in place — the group's designed day-two operation. An **empty list is valid**: create the anchor now, land the address plan later; until then rules referencing it match nothing.

**Name** -- the address set's INTENT ("branch-offices", "on-prem-ranges", "blocked-scanners"), unique within the resource group, up to 80 characters following ARM's Microsoft.Network naming contract. Renaming replaces the group, and every referencing rule must be re-pointed.

**Region** -- the group is a regional resource, but firewall policies in ANY region can reference it — one set can serve every regional firewall. Azure caps sharing at 100 firewall policies per group.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ip_group_id` | Azure Resource Manager ID of the group | Firewall policy rules (`source_ip_groups` / `destination_ip_groups`) and intrusion-detection traffic bypasses |
| `ip_group_name` | Name of the group | Automation scripts, inventory |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Branch offices** -- the trusted branch egress ranges every allow-rule references; a new branch is one address added here. Start from the **Branch-Office Address Set** preset.

**On-premises datacenter** -- the ExpressRoute/VPN-reachable ranges rules treat as internal — the hybrid-network trust anchor. Start from the **On-Premises Datacenter Ranges** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the IP Group is created
- [**Azure Firewall**](/infra-catalog/azure-firewall) -- the enforcement point whose policies reference this group's address set
- [**Azure Firewall Policy**](/infra-catalog/azure-firewall-policy) -- carries the rule collections whose rules reference this group by `ip_group_id`
- [**Azure Route Table**](/infra-catalog/azure-route-table) -- the complementary steering layer that sends traffic through the firewall these rules run on
