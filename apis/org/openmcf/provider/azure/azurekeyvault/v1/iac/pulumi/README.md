# Azure Key Vault Pulumi Module

## Overview

The **Azure Key Vault Pulumi Module** is part of the unified cloud-native API ecosystem developed to streamline multi-cloud infrastructure deployments. This module allows you to create and manage Azure Key Vault resources and secrets using a Kubernetes-like API resource model. The module is designed to work with Planton’s unified CLI, where users can easily define their desired infrastructure in an `api-resource.yaml` file and deploy it with minimal effort.

The key advantage of this module lies in its ability to seamlessly integrate with Azure’s Key Vault services and manage secrets based on the specifications in the API resource. Users can define multiple secret names to be created within the Key Vault, along with their respective values. The module interacts with Azure securely using provided credentials, and the results, including secret identifiers, are captured and exported in the `status.outputs`. This ensures a smooth and predictable infrastructure-as-code workflow, compatible with multi-cloud environments.

## Key Features

1. **Unified API Resource Model**: The module adheres to the standardized API resource model, which mimics Kubernetes resources with fields like `apiVersion`, `kind`, `metadata`, `spec`, and `status`. This provides a consistent experience across different cloud providers and services.

2. **Seamless Azure Integration**: The module sets up Azure Key Vault resources using Azure credentials passed through the `azure_credential_id`. It provisions Key Vaults and secrets with minimal configuration, making it easy for developers to manage sensitive data in a secure environment.

3. **Automated Secret Management**: Users can define a list of secret names, which are automatically created in the Azure Key Vault. The mapping between secret names and their corresponding IDs is captured in the output, ensuring that the necessary data is accessible post-deployment.

4. **Planton CLI Integration**: The module is fully integrated with Planton’s CLI, supporting the `planton pulumi up --stack-input <api-resource.yaml>` command. This ensures that users can deploy infrastructure by simply defining an API resource YAML file and executing the appropriate command. The CLI also supports specifying a Git repository for custom Pulumi modules, with a fallback to the default module created for each resource.

5. **Stack Outputs**: The outputs of the Pulumi module, such as secret IDs, are captured in the `status.outputs` field. This allows developers to programmatically reference these outputs for further automation or tracking purposes.

6. **Extensible and Scalable**: This module is built to scale, allowing users to manage numerous secrets across various environments, ensuring consistency and reducing the potential for manual errors.

## Usage

Refer to the example section for usage instructions.
