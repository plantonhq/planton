# Azure Resource Group

Deploys an Azure Resource Group in a specified region. Resource groups are the foundational organizational unit in Azure -- every other Azure resource must belong to one. This component creates the resource group and applies metadata tags for tracking and governance.

## What Gets Created

When you deploy an AzureResourceGroup resource, OpenMCF provisions:

- **Resource Group** — a `core.ResourceGroup` resource in the specified Azure region, serving as the container for all downstream Azure resources
- **Azure Tags** — metadata tags applied to the resource group including resource name, resource kind, organization, and environment for governance and cost tracking

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure subscription** where the resource group will be created
- **A unique name** for the resource group within the target subscription

## Quick Start

Create a file `resourcegroup.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: my-rg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureResourceGroup.my-rg
spec:
  name: my-rg
  region: eastus
```

Deploy:

```shell
openmcf apply -f resourcegroup.yaml
```

This creates a resource group named `my-rg` in the `eastus` region with standard metadata tags.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Name of the resource group. Must be unique within the Azure subscription. Allowed characters: alphanumeric, underscores, hyphens, periods, and parentheses. Cannot end with a period. | Required, 1–90 characters |
| `region` | `string` | Azure region where the resource group will be created (e.g., `eastus`, `westus2`, `westeurope`). Resources within the group can be in different regions, but the group's region determines where its metadata is stored. | Required, minimum length 1 |

### Optional Fields

This component has no optional fields. Resource groups are intentionally minimal containers -- they require only a name and a region.

## Examples

### Single Development Resource Group

A resource group for a development environment in the US East region:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: dev-rg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureResourceGroup.dev-rg
spec:
  name: dev-rg
  region: eastus
```

### Production Resource Group in Europe

A resource group for production workloads with metadata indicating the production environment:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: prod-eu-rg
  env: prod
  org: acme-corp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: acme-infra
    pulumi.openmcf.org/stack.name: prod.AzureResourceGroup.prod-eu-rg
spec:
  name: prod-eu-rg
  region: westeurope
```

### Multi-Region Resource Groups for DR

Multiple resource groups across regions, forming the basis of a disaster recovery topology. Each group hosts a replica of the application stack in its respective region.

Primary region:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: app-primary-rg
  env: prod
  org: acme-corp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: acme-infra
    pulumi.openmcf.org/stack.name: prod.AzureResourceGroup.app-primary-rg
spec:
  name: app-primary-rg
  region: eastus
```

Secondary region:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: app-secondary-rg
  env: prod
  org: acme-corp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: acme-infra
    pulumi.openmcf.org/stack.name: prod.AzureResourceGroup.app-secondary-rg
spec:
  name: app-secondary-rg
  region: westus2
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `resource_group_id` | `string` | Azure Resource Manager ID of the resource group (format: `/subscriptions/{subscription-id}/resourceGroups/{resource-group-name}`) |
| `resource_group_name` | `string` | Name of the resource group. This is the primary output referenced by downstream Azure resources via `StringValueOrRef` with `field: status.outputs.resource_group_name`. |
| `region` | `string` | Azure region where the resource group was created |

## Related Components

- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) — stores secrets, keys, and certificates within this resource group
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) — deploys a Kubernetes cluster into this resource group
- [AzureVpc](/docs/catalog/azure/azurevpc) — creates a Virtual Network within this resource group
- [AzureStorageAccount](/docs/catalog/azure/azurestorageaccount) — provisions blob, file, table, and queue storage in this resource group
- [AzurePostgresqlFlexibleServer](/docs/catalog/azure/azurepostgresqlflexibleserver) — deploys a PostgreSQL Flexible Server in this resource group
