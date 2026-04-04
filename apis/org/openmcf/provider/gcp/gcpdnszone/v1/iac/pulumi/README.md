# GCP DNS Zone Pulumi Module

## Overview

The **GCP DNS Zone Pulumi Module** is a robust solution designed to manage Google Cloud Platform (GCP) DNS Zones using a Kubernetes-inspired API resource model. This module streamlines the creation and management of DNS zones by leveraging the power of Pulumi and Go, enabling developers to define their DNS infrastructure as code within a familiar Kubernetes-like structure. By abstracting the complexities of direct GCP interactions, the module provides a standardized and scalable approach to DNS management in multi-cloud environments.

## Important Note

*This module is not completely implemented. Certain features may be missing or not fully functional. Future updates will address these limitations.*

## Key Features

### API Resource Features

- **Standardized Structure**: The `GcpDnsZone` API resource adheres to a consistent schema with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields. This uniformity ensures compatibility and ease of use within Kubernetes-like environments, facilitating seamless integration with existing workflows and tooling.
  
- **Configurable Specifications**:
  - **GCP Credential ID**: Securely reference the GCP credentials required to set up the Pulumi provider.
  - **Project ID**: Specifies the GCP project where the DNS Managed Zone will be created.
  - **IAM Service Accounts**: Define a list of GCP service accounts that are granted permissions to manage DNS records within the Managed Zone, facilitating secure and controlled access.
  - **DNS Records**: Comprehensive configuration of DNS records, including type, name, values, and TTL settings, allowing for precise DNS management.

- **Validation and Compliance**: Ensures that all DNS names follow valid formatting conventions and that required fields are properly validated, enhancing reliability and reducing configuration errors.

### Pulumi Module Features

- **Automated GCP Provider Setup**: Utilizes the provided GCP credentials to configure the Pulumi Google provider, enabling seamless interaction with GCP resources.
  
- **Managed Zone Creation**: Automatically creates a Managed Zone in GCP based on the provided specifications, handling necessary transformations such as replacing dots with hyphens in zone names to comply with GCP's naming requirements.

- **IAM Binding Configuration**: Establishes IAM bindings to grant specified service accounts the necessary permissions to manage DNS records within the zone. Although currently granting broader `dns.admin` roles at the project level due to provider limitations, the module is designed for future enhancements to support more granular permissions.

- **DNS Record Management**: Iterates through the defined DNS records in the API resource, creating each record within the Managed Zone with the appropriate type, name, values, and TTL. This ensures that all DNS configurations are accurately reflected in the GCP environment.

- **Exported Stack Outputs**: Provides essential outputs such as the Managed Zone name, nameservers, and project ID, which are captured in `status.outputs`. These outputs facilitate integration with other infrastructure components and enable effective status tracking within deployment workflows.

- **Error Handling**: Implements comprehensive error handling to ensure that any issues during the creation or configuration of resources are properly reported, aiding in troubleshooting and maintaining infrastructure integrity.

## Installation

To integrate the GCP DNS Zone Pulumi Module into your project, you can retrieve it from the [GitHub repository](https://github.com/your-repo/gcp-dns-zone-pulumi-module). Ensure that you have Pulumi and Go installed and properly configured in your development environment.

```shell
git clone https://github.com/your-repo/gcp-dns-zone-pulumi-module.git
cd gcp-dns-zone-pulumi-module
```

## Usage

Refer to the [example section](#examples) for usage instructions.

## Module Details

### Input Configuration

The module expects a `GcpDnsZoneStackInput` which includes:

- **Pulumi Input**: Configuration details required by Pulumi for managing the stack.
- **Target API Resource**: The `GcpDnsZone` resource defining the desired DNS zone configuration.
- **GCP Credential**: Specifications for the GCP credentials used to authenticate and authorize Pulumi operations.

### Exported Outputs

Upon successful execution, the module exports the following outputs to `status.outputs`:

- **Managed Zone Name**: The name of the created Managed Zone.
- **Nameservers**: The list of nameservers assigned to the Managed Zone.
- **GCP Project ID**: The ID of the GCP project where the Managed Zone resides.

These outputs enable seamless integration with other components of your infrastructure and provide essential information for monitoring and management purposes.

## Contributing

We welcome contributions to enhance the GCP DNS Zone Pulumi Module. Please refer to our [contribution guidelines](CONTRIBUTING.md) for more information on how to get involved.

## License

This project is licensed under the [MIT License](LICENSE).

## Support

For support, please contact our [support team](mailto:support@planton.ai).

## Acknowledgements

Special thanks to all contributors and the Planton community for their ongoing support and feedback.

## Changelog

Detailed changelog will be available in the [CHANGELOG.md](CHANGELOG.md) file.

## Roadmap

We are continuously working to enhance the GCP DNS Zone Pulumi Module. Upcoming features include:

- **Advanced IAM Configurations**: Implementing more granular permission controls for DNS resources.
- **Enhanced Monitoring Integrations**: Integrating with monitoring and logging tools for better observability.
- **Support for Additional Cloud Providers**: Extending support to more cloud platforms to increase flexibility and reach.

Stay tuned for more updates!

## Contact

For any inquiries or feedback, please reach out to us at [contact@planton.ai](mailto:contact@planton.ai).

## Disclaimer

*This project is maintained by Planton and is not affiliated with any third-party services unless explicitly stated.*

## Security

If you discover any security vulnerabilities, please report them responsibly by contacting our security team at [security@planton.ai](mailto:security@planton.ai).

## Code of Conduct

Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) when interacting with the project.

## References

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [GCP DNS Documentation](https://cloud.google.com/dns/docs)
- [Planton APIs](https://buf.build/openmcf/apis/docs)

## Getting Started

To get started with the GCP DNS Zone Pulumi Module, follow the installation instructions above and refer to the upcoming examples section for detailed usage guidelines.

---

*Thank you for choosing Planton's GCP DNS Zone Pulumi Module. We look forward to supporting your infrastructure management needs!*