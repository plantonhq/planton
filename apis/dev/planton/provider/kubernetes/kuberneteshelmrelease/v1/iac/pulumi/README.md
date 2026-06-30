# Helm Release Kubernetes Pulumi Module

## Key Features

- **Declarative Helm Management**: The module provides a declarative way to define and manage Helm releases through the `HelmRelease` API resource. Users can specify details like chart name, version, repository, and custom Helm values in a YAML configuration file.

- **Kubernetes Native Integration**: The module is fully integrated with Kubernetes, automating the provisioning of Helm releases and ensuring they are deployed to the correct Kubernetes cluster using the credentials provided in the API resource.

- **Pulumi-Driven Infrastructure Management**: Built on Pulumi’s infrastructure-as-code capabilities, the module ensures that any changes to the `HelmRelease` resource are reflected automatically in the Kubernetes cluster. Pulumi manages the lifecycle of the Helm release, from creation to updates and deletion.

- **Custom Helm Values**: The `HelmRelease` API resource allows users to pass custom key-value pairs as Helm chart values. This feature provides flexibility in configuring Helm charts according to specific application requirements.

- **Namespace Management**: The module supports both creating new namespaces and using existing ones based on the `create_namespace` flag. When set to `true`, a new namespace is created with appropriate labels. When set to `false`, the specified namespace name is used, assuming it already exists in the cluster.

- **Helm Chart Versioning**: Users can specify exact versions of Helm charts, ensuring compatibility and control over application deployments. This enables precise management of application lifecycle and infrastructure as versions of Helm charts change.

- **Environment-Specific Deployments**: The module integrates well with multi-environment setups. By specifying environment details in the API resource, Helm releases can be isolated and deployed in different environments, such as production, staging, or development, within the same Kubernetes cluster or across multiple clusters.

- **Output Management**: After deploying the Helm release, the module captures critical information such as:
  - The Kubernetes namespace where the Helm release is deployed.
  - Relevant details about the release, which can be used for monitoring or further configuration.

## Usage

To deploy Helm charts using this module, create a `HelmRelease` API resource in YAML format that specifies the desired Helm chart and configuration values. Once the YAML is created, you can deploy the Helm release in your Kubernetes cluster using the following command:

```bash
planton pulumi up --stack-input <api-resource.yaml>
```

Refer to the **Examples** section for detailed usage instructions.

## Pulumi Integration

This module is fully integrated with Pulumi’s Go SDK, enabling Helm release management as infrastructure code. By processing the `HelmRelease` API resource, the module provisions the necessary Kubernetes resources and Helm charts based on the specified configuration. Pulumi ensures that any updates to the Helm release configuration are reflected automatically in the cluster, simplifying Helm chart updates and management.

### Key Pulumi Components

1. **Kubernetes Provider**: The module configures the Kubernetes provider using the `kubernetes_credential_id` provided in the `HelmRelease` API resource. This ensures that all Helm releases are deployed in the correct Kubernetes cluster with the appropriate credentials.

2. **Namespace Management**: The module handles the creation or reuse of Kubernetes namespaces for the Helm release. It ensures that the release is properly isolated from other deployments in the cluster and stores the namespace information in the outputs.

3. **Helm Chart Management**: The module automates the installation and upgrade of Helm charts based on the `HelmRelease` specification. Users can define chart repositories, versions, and custom values to ensure the Helm release is configured according to the application's needs.

4. **Custom Helm Values**: The module allows passing custom key-value pairs to Helm charts. This is particularly useful for configuring Helm charts for different environments or fine-tuning application settings.

5. **Version Control**: By allowing users to specify the Helm chart version, the module ensures compatibility and provides precise control over the application lifecycle, allowing easy rollbacks or upgrades as necessary.

6. **Output Management**: Once the Helm release is deployed, the module captures essential details about the deployment, including the namespace where the release is created. These outputs can be used to monitor the release or configure additional resources.

## Status and Monitoring

All deployment outputs are captured in the `status.outputs` field, providing easy access to the namespace and other key information about the Helm release. This information is crucial for managing and monitoring the state of the Helm release in the Kubernetes cluster.

## Conclusion

The `helm-release-pulumi-module` provides a powerful, automated solution for managing Helm releases in Kubernetes. By using a declarative `HelmRelease` API resource and leveraging Pulumi’s infrastructure-as-code capabilities, the module simplifies Helm release management and ensures consistent deployments across different environments. With features like custom Helm values, version control, and namespace management, this module offers flexibility and control, making it suitable for both development and production environments.