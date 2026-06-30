# AWS Secrets Manager Pulumi Module

## Introduction

This Pulumi module automates the creation and management of secrets in AWS Secrets Manager using our Unified APIs that mimic Kubernetes' resource modeling. By utilizing this module, developers can define secrets to be managed in AWS Secrets Manager through a standardized YAML configuration, simplifying multi-cloud secrets management and infrastructure provisioning.

## Key Features

- **Unified API Structure**: Consistent resource modeling with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, similar to Kubernetes.
- **Multi-Cloud Support**: Designed to work seamlessly across multiple cloud providers, starting with AWS.
- **Pulumi Integration**: Leverages Pulumi's infrastructure-as-code capabilities to automate the provisioning and management of AWS Secrets Manager resources.
- **Credential Management**: Securely handles AWS credentials for authenticating with AWS services.
- **Simplified Secret Creation**: Allows developers to specify a list of secrets to be created in AWS Secrets Manager using a single YAML file.
- **Output Handling**: Captures and exports the ARNs of the created secrets, making them available in `status.outputs` for downstream consumption.
- **Standardized Documentation**: Comprehensive documentation available via buf.build for easy reference.

## Usage

Refer to the example section for usage instructions.

## Module Details

### API Resource Specification

The module expects an `api-resource.yaml` file defining the desired state of the AWS Secrets Manager resources. The key components of this file include:

- **`region`** (required): The AWS region where the resource will be created.
- **`secretNames`** (optional): A list of secret names to be created in AWS Secrets Manager.
- **`environmentInfo`**: Contains environment-specific information (can be ignored if not used).
- **`stackUpdateSettings`**: Settings related to the stack-update execution (can be ignored if not used).

### Pulumi Module Functionality

The core functionality of this module revolves around creating secrets in AWS Secrets Manager based on the provided specifications.

#### Steps Performed:

1. **AWS Provider Initialization**:  
   Initializes the AWS provider in Pulumi using credentials supplied in the `AwsSecretsManagerStackInput`. The credentials required are:

   - `AccessKeyId`
   - `SecretAccessKey`
   - `Region`

2. **Secret Creation**:  
   Iterates over each secret name specified in `spec.secretNames` and performs the following actions:

   - Constructs a unique secret ID by combining the resource's metadata ID and the secret name.
   - Creates a secret in AWS Secrets Manager with a placeholder value.
   - Sets appropriate tags for resource identification and management.

3. **Output Handling**:  
   Exports the ARN of each created secret, capturing them in `status.outputs` for easy reference by other resources or applications.

## Limitations

- **Placeholder Secret Values**: The secrets are created with a placeholder value. Updating the secret value is not handled by this module and should be managed separately.
- **No Secret Value Management**: This module does not support setting or updating the actual secret values beyond the initial placeholder.
- **Limited Error Handling**: While basic error checks are in place, comprehensive error handling and input validation may need enhancement.

## Future Enhancements

- **Secret Value Management**: Implement functionality to securely set and update secret values.
- **Advanced Error Handling**: Introduce more robust error handling and input validation to improve reliability.
- **Environment and Stack Job Settings**: Utilize `environmentInfo` and `stackUpdateSettings` to allow for more granular control over deployment environments and execution settings.
- **Multi-Region Support**: Extend the module to support creating secrets across multiple AWS regions.

## Documentation

For detailed API definitions and additional documentation, please refer to our resources available via [buf.build](https://buf.build).

## Contributing

Contributions are welcome! Please open issues or pull requests to help improve this module.

## License

This project is licensed under the [MIT License](LICENSE).
