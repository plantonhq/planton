# Azure AKS Cluster Pulumi Module

## Introduction

This Pulumi module provides a standardized way to manage Azure Kubernetes Service (AKS) clusters using our Unified APIs that mimic Kubernetes' resource modeling. It allows developers to define infrastructure configurations in a YAML file, simplifying the deployment and management of complex infrastructure across multiple providers.

## Key Features

- **Unified API Structure**: Adheres to a standardized API format with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, ensuring consistency across different resources.
- **Multi-Cloud Support**: Designed to work seamlessly in a multi-cloud environment, starting with Azure.
- **Pulumi Integration**: Leverages Pulumi's infrastructure-as-code capabilities to automate resource provisioning.
- **Credential Management**: Securely handles Azure credentials for authenticating with Azure services.
- **Simplified Deployment**: Enables developers to deploy AKS clusters using a single YAML configuration file.
- **Standardized Documentation**: Comprehensive documentation available via buf.build for easy reference.

## Usage

Deploy from a manifest with the planton CLI. From this directory (it holds `Pulumi.yaml`, so the CLI runs this module):

```bash
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p azure-provider-config.yaml
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p azure-provider-config.yaml
```

The CLI builds the `AzureAksClusterIacInput` (the manifest as `target`, the `-p` file as `provider_config`) and hands it to the module through `IAC_INPUT_YAML_FILE`. The module also reads that input from the Pulumi config key `planton:iac-input` or from `IAC_INPUT_YAML` (YAML content).

## Module Details

### API Resource Specification

The module reads an `AzureAksClusterIacInput` with two fields:

- **`target`** (required): the `AzureAksCluster` manifest. Its `spec` carries, among others:
  - **`resource_group`** (required): the resource group, as a literal name or a reference to an `AzureResourceGroup`
  - **`region`**, **`name`** (required), **`dns_prefix`**, **`kubernetes_version`**, **`sku_tier`**, **`support_plan`**
  - **`default_node_pool`** (required): the system pool; every other pool is a separate `AzureAksNodePool`
  - **`identity`**, **`kubelet_identity`**, **`oidc_issuer_enabled`**, **`workload_identity_enabled`**
  - **`private_cluster_enabled`**, **`api_server_access_profile`**, **`network_profile`**, Entra ID RBAC, upgrade channels, maintenance windows, and add-ons
- **`provider_config`**: the Azure credentials (an `AzureProviderConfig`)

### Pulumi Module Functionality

The module sets up the Azure provider from the supplied credentials, then provisions the resource from the spec and exports its outputs.

#### Steps Performed:

1. **Azure Provider Initialization**:  
   Initializes the Azure provider from the `provider_config` in the `AzureAksClusterIacInput`:

   - `client_id`, `tenant_id`, `subscription_id` (required)
   - `client_secret` for a service principal, or `web_identity.web_identity_token` for keyless federation; with neither, the provider uses the ambient Azure credential chain

2. **Resource Provisioning**:  
   Creates the AKS cluster with its default node pool, identity, network profile, maintenance windows, add-ons, and platform profiles from the spec.

3. **Output Handling**:  
   Exports `cluster_id`, `cluster_name`, `fqdn`, `private_fqdn`, `portal_fqdn`, `oidc_issuer_url`, `node_resource_group`, `node_resource_group_id`, `cluster_kubeconfig`, `cluster_ca_certificate`, `cluster_identity_principal_id`, `kubelet_identity_object_id`, `kubelet_identity_client_id`, `current_kubernetes_version`, and `entra_integration_enabled`; Planton stores them in `status.outputs`.

## Limitations

- **One system pool**: the cluster carries only its default node pool; add every other pool as a separate `AzureAksNodePool` resource.
- **Resource group changes replace the cluster**: changing `resource_group` creates a new cluster.

## Documentation

For detailed API definitions and additional documentation, please refer to our resources available via [buf.build](https://buf.build).

## Contributing

Contributions are welcome! Please open issues or pull requests to help improve this module.

## License

This project is licensed under the [MIT License](LICENSE).
