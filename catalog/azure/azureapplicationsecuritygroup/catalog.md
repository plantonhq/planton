# Azure Application Security Group

Deploys an Azure Application Security Group (ASG) — a named, logical grouping of network interfaces that NSG rules can target by name instead of by IP address. An ASG lets a security policy say "allow the web tier to reach the app tier" rather than "allow 10.0.1.0/24 to reach 10.0.2.0/24" — the rule follows the workload as instances scale in and out and change addresses, so micro-segmentation stops being tied to a fragile CIDR plan. The group itself is **deliberately empty**: membership is declared from the member's side — a network interface lists the ASGs it joins, and NSG security rules reference source/destination groups.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Application Security Group** -- the named grouping anchor, holding no members and no rules of its own
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

Memberships and rules are NOT created here — a NIC joins from its own spec (`applicationSecurityGroupIds`), and NSG rules target the group from theirs. That inversion is what makes the ASG a stable composition anchor: created once, referenced by many, each with its own lifecycle.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the group will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A naming plan**: one group per workload ROLE (web, app-tier, db) — the names become the vocabulary your NSG rules are written in.

## Deploy

### Console

Open the deployment store, find **Azure Application Security Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Workload Tier Group** preset in the [Presets](#presets) tab for the classic role-named group.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureApplicationSecurityGroup
metadata:
  name: web-tier
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-network-prod"
  name: web-tier
  tags:
    cost-center: platform-network
```

```shell
planton apply -f asg.yaml
```

This creates one empty role-named group ready for NICs to join and NSG rules to target — create one per tier and the security policy reads as intent, surviving every scale event without a firewall rewrite. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the group to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-prod
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the group — and the NICs that join it reference this group's `application_security_group_id`.

## Key Configuration

These are the most important decisions when configuring an Application Security Group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Name** -- the workload role's identity ("web", "app-tier", "db"), unique within the resource group, up to 80 characters following ARM's Microsoft.Network naming contract. The name IS the security vocabulary — NSG rules read it as intent — so name the role, never a machine. Renaming replaces the group, and every rule and NIC that referenced it must be re-pointed.

**Region** -- an ASG can only be joined by network interfaces in the same region. A multi-region application stamps the same tier groups per region so the rule vocabulary stays identical everywhere.

**Tags** -- the ONLY field on an ASG that updates in place; everything else is fixed at creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `application_security_group_id` | Azure Resource Manager ID of the group | AzureNetworkInterface `applicationSecurityGroupIds` (membership), NSG rule source/destination groups (targeting) |
| `application_security_group_name` | Name of the group | Automation scripts, inventory |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload tier** -- one group per application tier ("web", "app-tier", "db"); each tier's NICs join their group and the NSG rules read as an architecture diagram. Start from the **Workload Tier Group** preset.

**Tagged governance** -- the same role-named group with the ownership tags your Azure Policy regime enforces, for organizations that gate deployments on tag compliance. Start from the **Governed Data-Tier Group** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the ASG is created
- [**Azure Network Interface**](/infra-catalog/azure-network-interface) -- joins the group by referencing its `application_security_group_id` output (membership lives on the NIC)
- [**Azure Network Security Group**](/infra-catalog/azure-network-security-group) -- targets the group in security rules as source/destination, turning CIDR rules into role rules
- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the workload whose NICs carry the memberships that make the group mean something
