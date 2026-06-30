**Note:** This module is not completely implemented as the API resource specification is currently empty.

# Azure Azure AKS Cluster Pulumi Module

## Introduction

This Pulumi module provides a standardized way to manage Azure Kubernetes Service (AKS) clusters using our Unified APIs that mimic Kubernetes' resource modeling. It allows developers to define infrastructure configurations in a YAML file, simplifying the deployment and management of complex cloud resources across multiple providers.

## Key Features

- **Unified API Structure**: Adheres to a standardized API format with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, ensuring consistency across different resources.
- **Multi-Cloud Support**: Designed to work seamlessly in a multi-cloud environment, starting with Azure.
- **Pulumi Integration**: Leverages Pulumi's infrastructure-as-code capabilities to automate resource provisioning.
- **Credential Management**: Securely handles Azure credentials for authenticating with Azure services.
- **Simplified Deployment**: Enables developers to deploy AKS clusters using a single YAML configuration file.
- **Standardized Documentation**: Comprehensive documentation available via buf.build for easy reference.

## Usage

Refer to the example section for usage instructions.

## Module Details

### API Resource Specification

The module expects an `api-resource.yaml` file defining the desired state of the AKS cluster. The key components of this file include:

- **`azure_credential_id`** (required): The identifier for the Azure credentials used to authenticate with Azure services.
- **`environment_info`**: Contains environment-specific information (currently not implemented).
- **`stack_job_settings`**: Settings related to the stack-update execution (currently not implemented).

### Pulumi Module Functionality

The core functionality of this module revolves around setting up the Azure provider within the Pulumi context using the provided Azure credentials. This setup is essential for any subsequent resource creation and management within Azure.

#### Steps Performed:

1. **Azure Provider Initialization**:  
   Initializes the Azure provider in Pulumi using credentials supplied in the `AzureAksNodePoolStackInput`. The credentials required are:

   - `ClientId`
   - `ClientSecret`
   - `SubscriptionId`
   - `TenantId`

2. **Resource Provisioning**:  
   *(Not yet implemented)* The module will provision the AKS cluster and any associated resources based on the specifications provided in the `api-resource.yaml` file.

3. **Output Handling**:  
   *(Not yet implemented)* Captures the outputs from the Pulumi stack execution and stores them in `status.outputs` for later reference.

## Limitations

- **Incomplete Implementation**: The module currently does not implement resource creation due to the empty API resource specification.
- **Unused Spec Fields**: Fields like `environment_info` and `stack_job_settings` are included in the spec but are not utilized in the current implementation.
- **No Error Handling**: Advanced error handling and validation mechanisms are yet to be implemented.

## Future Enhancements

- **Implement Resource Creation**: Extend the module to create AKS clusters and related Azure resources based on the provided specifications.
- **Utilize Spec Fields**: Make use of `environment_info` and `stack_job_settings` to allow for more granular control over the deployment environment and stack-update configurations.
- **Enhance Output Management**: Capture and expose essential output parameters such as cluster endpoints, credentials, and configuration details.
- **Error Handling and Validation**: Introduce comprehensive error handling and input validation to improve reliability and user experience.

## Documentation

For detailed API definitions and additional documentation, please refer to our resources available via [buf.build](https://buf.build).

## Contributing

Contributions are welcome! Please open issues or pull requests to help improve this module.

## License

This project is licensed under the [MIT License](LICENSE).
