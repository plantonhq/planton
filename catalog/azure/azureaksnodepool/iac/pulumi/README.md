# Azure AKS Node Pool Pulumi Module

## Introduction

This Pulumi module provides a standardized way to manage Azure Kubernetes Service (AKS) node pools using our Unified APIs that mimic Kubernetes' resource modeling. It allows developers to define infrastructure configurations in a YAML file, simplifying the deployment and management of complex infrastructure across multiple providers.

## Key Features

- **Unified API Structure**: Adheres to a standardized API format with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, ensuring consistency across different resources.
- **Multi-Cloud Support**: Designed to work seamlessly in a multi-cloud environment, starting with Azure.
- **Pulumi Integration**: Leverages Pulumi's infrastructure-as-code capabilities to automate resource provisioning.
- **Credential Management**: Securely handles Azure credentials for authenticating with Azure services.
- **Simplified Deployment**: Enables developers to add node pools to AKS clusters using a single YAML configuration file.
- **Standardized Documentation**: Comprehensive documentation available via buf.build for easy reference.

## Usage

Deploy from a manifest with the planton CLI. From this directory (it holds `Pulumi.yaml`, so the CLI runs this module):

```bash
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p azure-provider-config.yaml
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p azure-provider-config.yaml
```

The CLI builds the `AzureAksNodePoolIacInput` (the manifest as `target`, the `-p` file as `provider_config`) and hands it to the module through `IAC_INPUT_YAML_FILE`. The module also reads that input from the Pulumi config key `planton:iac-input` or from `IAC_INPUT_YAML` (YAML content).

## Module Details

### API Resource Specification

The module reads an `AzureAksNodePoolIacInput` with two fields:

- **`target`** (required): the `AzureAksNodePool` manifest. Its `spec` carries, among others:
  - **`kubernetes_cluster_id`** (required): the parent AKS cluster's ARM ID, as a literal or a reference to an `AzureAksCluster`'s `status.outputs.cluster_id`
  - **`name`**, **`vm_size`** (required), **`mode`**, **`os_type`**, **`os_sku`**
  - **`node_count`**, **`auto_scaling_enabled`**, **`min_count`**, **`max_count`**, **`max_pods`**
  - **`priority`**, **`eviction_policy`**, **`spot_max_price`** for Spot pools
  - **`node_labels`**, **`node_taints`**, **`zones`**, **`vnet_subnet_id`**, **`pod_subnet_id`**, disk, GPU, upgrade, kubelet, and Linux OS settings
- **`provider_config`**: the Azure credentials (an `AzureProviderConfig`)

### Pulumi Module Functionality

The module sets up the Azure provider from the supplied credentials, then provisions the resource from the spec and exports its outputs.

#### Steps Performed:

1. **Azure Provider Initialization**:  
   Initializes the Azure provider from the `provider_config` in the `AzureAksNodePoolIacInput`:

   - `client_id`, `tenant_id`, `subscription_id` (required)
   - `client_secret` for a service principal, or `web_identity.web_identity_token` for keyless federation; with neither, the provider uses the ambient Azure credential chain

2. **Resource Provisioning**:  
   Creates the node pool on the cluster named by `kubernetes_cluster_id`, deriving the resource group and cluster name from that ARM ID.

3. **Output Handling**:  
   Exports `node_pool_id`, `node_pool_name`, and `node_image_version`; Planton stores them in `status.outputs`.

## Limitations

- **Cluster changes replace the pool**: changing `kubernetes_cluster_id` creates a new pool.
- **Renames replace the pool**: changing `name` replaces the pool unless `temporary_name_for_rotation` is set.

## Documentation

For detailed API definitions and additional documentation, please refer to our resources available via [buf.build](https://buf.build).

## Contributing

Contributions are welcome! Please open issues or pull requests to help improve this module.

## License

This project is licensed under the [MIT License](LICENSE).
