# Overview

The **Azure Container Registry API Resource** provides a consistent and standardized interface for deploying and managing Azure Container Registry (ACR) instances within your infrastructure. This resource simplifies the provisioning of private Docker/OCI container registries on Azure, enabling secure storage and distribution of container images.

## Purpose

We developed this API resource to streamline the deployment and management of Azure Container Registry. By offering a unified interface, it reduces the complexity involved in setting up private container registries on Azure, enabling you to:

- **Easily Deploy Container Registries**: Quickly provision ACR instances with minimal configuration
- **Choose the Right SKU**: Select from Basic, Standard, or Premium tiers based on your requirements
- **Enable Geo-Replication**: Distribute images across multiple Azure regions for global deployments (Premium)
- **Integrate with AKS**: Seamlessly connect registries to Azure Kubernetes Service clusters
- **Control Access**: Configure authentication using Azure AD, managed identities, or admin credentials

## Key Features

- **Consistent Interface**: Aligns with Planton's APIs for deploying cloud infrastructure and services
- **Simplified Deployment**: Automates the provisioning of ACR, including resource groups and SKU configuration
- **SKU Flexibility**: Support for Basic, Standard, and Premium tiers to match performance and cost requirements
- **Geo-Replication**: Automatic setup of multi-region replicas for Premium SKUs
- **Security-First**: Admin user disabled by default; supports Azure AD integration and managed identities
- **Integration**: Works seamlessly with Azure Kubernetes Service, Azure DevOps, and GitHub Actions

## Use Cases

- **Private Container Images**: Store and manage private Docker images for your applications
- **AKS Integration**: Provide container images to Azure Kubernetes Service clusters
- **CI/CD Pipelines**: Use as a target for Docker builds in Azure DevOps or GitHub Actions
- **Global Deployments**: Leverage geo-replication for multi-region Kubernetes deployments
- **Development and Production**: Use Basic SKU for dev/test, Standard/Premium for production

## SKU Tiers

- **Basic**: Development and testing (~$5/month, 10 GB storage)
- **Standard**: Production workloads (~$20/month, 100 GB storage, higher throughput)
- **Premium**: Enterprise features (~$20/month base + replica costs, geo-replication, private link)

## Minimal Example

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerRegistry
metadata:
  name: my-container-registry
spec:
  region: eastus
  registryName: mycompanyacr
```

This creates a Standard SKU container registry in East US region with admin user disabled (production best practice).

## Next Steps

- Review [examples.md](./examples.md) for detailed configuration examples
- Read the [research documentation](./docs/README.md) for deployment best practices
- Check the [Pulumi module documentation](./iac/pulumi/README.md) for IaC details
