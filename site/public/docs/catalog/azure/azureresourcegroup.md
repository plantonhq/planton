---
title: "Resourcegroup"
description: "Resourcegroup deployment documentation"
icon: "package"
order: 100
componentName: "azureresourcegroup"
---

# AzureResourceGroup: Research & Design Documentation

## 1. What Is an Azure Resource Group?

An Azure Resource Group is a logical container that holds related Azure resources for
a solution. It is the fundamental unit of organization, lifecycle management, access
control, and cost tracking in Azure.

Every Azure resource -- from a virtual machine to a storage account to a Kubernetes
cluster -- must belong to exactly one resource group. A resource group cannot contain
another resource group (they are flat, not hierarchical), and a resource cannot span
multiple resource groups.

### Key Properties

- **Region**: A resource group has a location (region) that determines where its
  metadata is stored. Resources within the group can be in different regions.
- **Lifecycle**: Deleting a resource group deletes all resources within it. This
  makes it a powerful cleanup mechanism for ephemeral environments.
- **RBAC boundary**: Azure Role-Based Access Control (RBAC) can be scoped to a
  resource group, enabling fine-grained permission management.
- **Cost boundary**: Azure Cost Management and billing can report costs per resource
  group, enabling cost allocation to teams or projects.
- **Tag inheritance**: Tags applied to a resource group are NOT inherited by child
  resources (a common misconception). Each resource must be tagged individually.

## 2. Deployment Landscape

### How People Deploy Resource Groups Today

#### Level 0: Azure Portal (Click-Ops)

Most Azure users start by creating resource groups through the Azure Portal. This is
fine for learning but creates undocumented, unreproducible infrastructure.

#### Level 1: Azure CLI

```bash
az group create --name my-rg --location eastus --tags env=production
```

Simple and scriptable, but lacks state management and drift detection.

#### Level 2: ARM Templates / Bicep

```bicep
targetScope = 'subscription'

resource rg 'Microsoft.Resources/resourceGroups@2021-04-01' = {
  name: 'my-rg'
  location: 'eastus'
  tags: {
    env: 'production'
  }
}
```

Azure-native IaC with full resource group lifecycle management.

#### Level 3: Terraform

```hcl
resource "azurerm_resource_group" "main" {
  name     = "my-rg"
  location = "eastus"
  tags     = { env = "production" }
}
```

The most popular approach for multi-cloud teams. Terraform manages resource group
state, enables plan/apply workflows, and integrates with CI/CD pipelines.

#### Level 4: Pulumi

```go
rg, _ := core.NewResourceGroup(ctx, "my-rg", &core.ResourceGroupArgs{
    Name:     pulumi.String("my-rg"),
    Location: pulumi.String("eastus"),
    Tags:     pulumi.ToStringMap(map[string]string{"env": "production"}),
})
```

Programmatic IaC using real programming languages. Preferred for teams that want
type safety, testing, and abstraction capabilities.

#### Level 5: OpenMCF (This Component)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: my-rg
spec:
  name: my-rg
  region: eastus
```

Declarative, Kubernetes-style API that abstracts Pulumi/Terraform behind a consistent
multi-cloud interface. Enables infra chart composition where resource groups become
Layer 0 nodes in the deployment DAG.

## 3. Why OpenMCF Models Resource Groups

### The Composability Argument

OpenMCF's power comes from `StringValueOrRef` -- the ability for one resource to
reference another resource's output. This creates the dependency graph that infra
charts use to deploy resources in the correct order.

Without a first-class resource group resource, the dependency graph has a gap:

```
                 [???]                    <- No resource group node
                /     \
     [Key Vault]      [Storage Account]   <- These need a resource group
```

With a first-class resource group:

```
          [AzureResourceGroup]            <- Layer 0: foundation
           /        |        \
  [KeyVault]  [StorageAccount]  [LAW]     <- Layer 1: reference via valueFrom
```

The platform can now:
- Visualize the full dependency graph
- Perform impact analysis ("what happens if I delete this resource group?")
- Deploy in correct topological order (resource group first)
- Track costs per resource group in the topology

### The 80/20 Principle Applied

Resource groups are perhaps the purest example of the 80/20 principle in Azure. The
`azurerm_resource_group` Terraform resource has only three attributes:

1. `name` (required)
2. `location` (required)
3. `tags` (optional, handled by OpenMCF metadata)

There are no optional features, no SKU tiers, no complex configuration blocks. The
resource group exists solely to contain other resources. Our spec mirrors this
simplicity exactly.

## 4. Resource Group Naming Conventions

### Azure Constraints

- 1 to 90 characters
- Alphanumeric, underscores, hyphens, periods, and parentheses
- Cannot end with a period
- Must be unique within the subscription

### Common Enterprise Patterns

| Pattern | Example | Use Case |
|---------|---------|----------|
| `{env}-{app}-rg` | `prod-platform-rg` | Application-scoped |
| `{env}-{tier}-rg` | `prod-network-rg` | Tier-scoped (networking, database, app) |
| `{org}-{env}-{region}-rg` | `contoso-prod-eastus-rg` | Multi-region enterprises |
| `rg-{app}-{env}` | `rg-platform-prod` | Azure CAF (Cloud Adoption Framework) |

### Recommendation

Use `{env}-{purpose}-rg` for most cases. It's readable, sortable by environment, and
clearly communicates intent.

## 5. Resource Group vs. Subscription vs. Management Group

Azure has a hierarchy of organizational constructs:

```
Management Group (optional)
  └── Subscription
       └── Resource Group
            └── Resources
```

- **Management Groups**: Group subscriptions for policy inheritance. Rarely managed
  via IaC in most organizations.
- **Subscriptions**: Billing and policy boundaries. Usually pre-provisioned by IT.
  Not a typical IaC target for application teams.
- **Resource Groups**: The primary IaC target for application infrastructure. This
  is where OpenMCF operates.

## 6. Design Decisions

### Why `name` Is a Spec Field (Not Derived from Metadata)

In most OpenMCF resources, the resource name is derived from `metadata.name`.
For resource groups, we expose `name` as an explicit spec field because:

1. Resource group names follow enterprise naming conventions that may differ
   from the OpenMCF metadata name
2. Multiple OpenMCF resources may reference the same resource group by name
3. The resource group name is an Azure-level identifier that appears in Azure
   Portal, CLI output, and ARM resource IDs

The `metadata.name` serves as the OpenMCF-level identifier for DAG wiring,
while `spec.name` is the Azure-level resource group name.

### Why No Tags in Spec

Tags are handled automatically via OpenMCF metadata (name, org, env) and
propagated to Azure tags in the IaC modules. This follows the established
pattern across all OpenMCF Azure resources.

If users need additional custom tags beyond what metadata provides, this can
be added as an optional `map<string, string> tags` field in a future iteration.

## 7. Infra Chart Integration

### As Layer 0

In every Azure infra chart, resource group is the root node:

```
AzureResourceGroup (Layer 0)
├── AzureVpc (Layer 1)
│   ├── AzureSubnet (Layer 2)
│   │   ├── AzurePostgresqlFlexibleServer (Layer 3)
│   │   └── AzureContainerAppEnvironment (Layer 3)
│   └── AzureNatGateway (Layer 2)
├── AzureLogAnalyticsWorkspace (Layer 1)
│   └── AzureApplicationInsights (Layer 2)
└── AzureKeyVault (Layer 1)
```

### ValueFrom Wiring

Every downstream Azure resource uses:

```yaml
resource_group:
  valueFrom:
    kind: AzureResourceGroup
    name: "{{ values.env }}-platform-rg"
    fieldPath: status.outputs.resource_group_name
```

This creates the DAG edge from the resource group to the dependent resource.

## 8. Scope Boundaries

### What This Component Does

- Creates an Azure Resource Group with the specified name and region
- Tags the resource group with OpenMCF metadata
- Exports the resource group ID, name, and region for downstream consumption

### What This Component Does NOT Do

- **Resource group locks** -- Azure resource locks (CanNotDelete, ReadOnly) are
  a governance concern handled at the subscription/policy level
- **RBAC assignments** -- role assignments are managed separately via
  AzureUserAssignedIdentity or Azure Policy
- **Budget alerts** -- cost management is an Azure-level feature, not resource-level
- **Policy assignments** -- Azure Policy is a subscription/management group concern
- **Resource group moves** -- moving resources between resource groups is an
  operational task, not an IaC concern
