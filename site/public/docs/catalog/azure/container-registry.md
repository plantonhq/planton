---
title: "Container Registry"
description: "Container Registry deployment documentation"
icon: "package"
order: 100
componentName: "azurecontainerregistry"
---

# Azure Container Registry

Deploys an Azure Container Registry with a configurable SKU tier, optional admin user access, and geo-replication to additional regions. The component provisions a single registry resource and, for Premium SKU deployments, creates replication resources for each specified region.

## What Gets Created

When you deploy an AzureContainerRegistry resource, OpenMCF provisions:

- **Container Registry** — a `containerregistry.Registry` resource in the specified region and resource group, configured with the chosen SKU tier, admin user setting, and network rule bypass for Azure services
- **Geo-Replications** — for Premium SKU only, a `containerregistry.Replication` resource for each entry in `geoReplicationRegions`, enabling multi-region image pull performance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the registry will be created (can reference an AzureResourceGroup resource)
- **A globally unique registry name** — must be 5-50 characters of lowercase letters or numbers, unique across all of Azure

## Quick Start

Create a file `acr.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerRegistry
metadata:
  name: my-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureContainerRegistry.my-registry
spec:
  region: eastus
  resourceGroup: my-rg
  registryName: myregistry01
```

Deploy:

```shell
openmcf apply -f acr.yaml
```

This creates a Standard-tier container registry with admin user disabled and network rule bypass configured for Azure trusted services.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the container registry (e.g., `eastus`, `westeurope`). | Required |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `registryName` | `string` | Globally unique name for the container registry. | Required, 5-50 lowercase alphanumeric characters (`^[a-z0-9]{5,50}$`) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sku` | `enum` | `STANDARD` | Container registry pricing tier. Values: `BASIC` (cost-effective for development), `STANDARD` (production workloads with higher throughput), `PREMIUM` (geo-replication, content trust, and private link support). |
| `adminUserEnabled` | `bool` | `false` | Enables the admin user account for the registry. Use only for basic authentication scenarios; service principals or managed identities are recommended instead. |
| `geoReplicationRegions` | `string[]` | `[]` | Additional Azure regions to replicate the registry for low-latency pulls. Only applicable when `sku` is `PREMIUM`. |

## Examples

### Development Registry

A minimal registry for development with the lowest-cost SKU:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerRegistry
metadata:
  name: dev-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureContainerRegistry.dev-registry
spec:
  region: eastus
  resourceGroup: dev-rg
  registryName: devregistry01
  sku: BASIC
  adminUserEnabled: true
```

### Standard Production Registry

A production registry using the Standard tier with admin user disabled:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerRegistry
metadata:
  name: prod-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerRegistry.prod-registry
spec:
  region: eastus
  resourceGroup: prod-rg
  registryName: prodregistry01
  sku: STANDARD
```

### Premium Registry with Geo-Replication

A Premium-tier registry replicated across multiple regions for low-latency image pulls in a globally distributed deployment:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerRegistry
metadata:
  name: global-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerRegistry.global-registry
spec:
  region: eastus
  resourceGroup: prod-rg
  registryName: globalregistry01
  sku: PREMIUM
  geoReplicationRegions:
    - westeurope
    - southeastasia
    - westus2
```

### Using Foreign Key References

Reference an OpenMCF-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureContainerRegistry
metadata:
  name: ref-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureContainerRegistry.ref-registry
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  registryName: refregistry01
  sku: PREMIUM
  geoReplicationRegions:
    - westeurope
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `registryLoginServer` | `string` | The registry's login server URL for pulling and pushing images (e.g., `myregistry.azurecr.io`). |
| `registryResourceId` | `string` | Azure Resource Manager ID of the container registry. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) — provides the resource group for registry placement
- [AzureAksCluster](/docs/catalog/azure/aks-cluster) — AKS clusters pull container images from the registry
- [AzureKeyVault](/docs/catalog/azure/key-vault) — stores registry admin credentials or service principal secrets used for authentication
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) — provides VNet subnets for private endpoint connectivity to Premium-tier registries
