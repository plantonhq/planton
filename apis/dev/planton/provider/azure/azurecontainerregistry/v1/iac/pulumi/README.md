# Azure Container Registry - Pulumi Module

This Pulumi module deploys an Azure Container Registry (ACR) instance using the Azure Native provider.

## Overview

The module creates:
- **Resource Group**: Dedicated resource group for the container registry
- **Container Registry**: ACR instance with configurable SKU (Basic, Standard, Premium)
- **Geo-Replications**: Optional multi-region replicas for Premium SKU

## Features

- **Multiple SKU Support**: Basic, Standard, and Premium tiers
- **Geo-Replication**: Automatic setup of regional replicas for Premium
- **Secure by Default**: Admin user disabled unless explicitly enabled
- **Azure AD Integration**: Native support for managed identities and service principals
- **Network Security**: Configurable network rules (handled via Azure-native settings)

## Usage

This module is typically invoked through Planton's orchestration layer. For standalone usage:

```go
import (
    azurecontainerregistryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurecontainerregistry/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurecontainerregistry/v1/iac/pulumi/module"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &azurecontainerregistryv1.AzureContainerRegistryStackInput{
            Target: &azurecontainerregistryv1.AzureContainerRegistry{
                Metadata: &shared.CloudResourceMetadata{
                    Name: "my-acr",
                },
                Spec: &azurecontainerregistryv1.AzureContainerRegistrySpec{
                    Region:           "eastus",
                    RegistryName:     "mycompanyacr",
                    Sku:              azurecontainerregistryv1.AzureContainerRegistrySku_STANDARD,
                    AdminUserEnabled: false,
                },
            },
            ProviderConfig: &azureproviderv1.AzureProviderConfig{
                // Provider credentials
            },
        }
        return module.Resources(ctx, stackInput)
    })
}
```

## Inputs

See `spec.proto` for the complete specification. Key fields:

- **region** (required): Azure region (e.g., "eastus")
- **registry_name** (required): Globally unique registry name (5-50 lowercase alphanumeric)
- **sku**: SKU tier (BASIC, STANDARD, PREMIUM)
- **admin_user_enabled**: Enable admin user (default: false)
- **geo_replication_regions**: List of regions for Premium geo-replication

## Outputs

- **registry_login_server**: Login server URL (e.g., "myacr.azurecr.io")
- **registry_resource_id**: Azure Resource Manager ID

## SKU Selection

- **Basic**: Dev/test only (~$5/month, 10 GB)
- **Standard**: Production default (~$20/month, 100 GB)
- **Premium**: Enterprise with geo-replication (~$20/month + $50/replica)

## Geo-Replication

Only available with Premium SKU. Specify additional regions to replicate:

```yaml
sku: PREMIUM
geoReplicationRegions:
  - westeurope
  - southeastasia
```

Azure automatically routes pulls to the nearest replica.

## Authentication

### From Azure CLI
```shell
az acr login --name mycompanyacr
```

### From AKS (Managed Identity)
```shell
az aks update \
  --name my-cluster \
  --resource-group rg-my-cluster \
  --attach-acr mycompanyacr
```

### From CI/CD
Use Azure service principal or GitHub OIDC with AcrPush/AcrPull roles.

## Debugging

Use the provided debug script:

```shell
cd iac/pulumi
./debug.sh
```

## Module Structure

- `main.go`: Core resource provisioning logic
- `outputs.go`: Output constants matching stack_outputs.proto
- `locals.go`: (Optional) Computed values and transformations

## References

- [Azure Container Registry Documentation](https://docs.microsoft.com/en-us/azure/container-registry/)
- [Pulumi Azure Native Provider](https://www.pulumi.com/registry/packages/azure-native/)
- [Research Documentation](../../docs/README.md)
