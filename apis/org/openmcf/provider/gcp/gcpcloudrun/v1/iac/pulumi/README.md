# GCP Cloud Run Pulumi Module

## Key Features

### API Resource Features

- **Standardized Structure**: The `GcpCloudRun` API resource adheres to a consistent schema with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields. This uniformity ensures compatibility and ease of integration within Kubernetes-like environments, promoting seamless workflow incorporation and tooling interoperability.
  
- **Configurable Specifications**:
  - **GCP Credential ID**: Securely references the GCP credentials required to set up the Pulumi provider, ensuring authenticated and authorized interactions with GCP services.
  - **GCP Project ID**: Specifies the GCP project where the Cloud Run resources will be deployed, enabling precise management and organization of cloud resources.

- **Validation and Compliance**: Incorporates stringent validation rules to ensure all configurations adhere to established standards and best practices, minimizing the risk of misconfigurations and enhancing overall system reliability.

### Pulumi Module Features

- **Automated GCP Provider Setup**: Leverages the provided GCP credentials to automatically configure the Pulumi GCP provider, enabling seamless and secure interactions with GCP Cloud Run resources.
  
- **Cloud Run Management**: Streamlines the creation and management of GCP Cloud Run configurations based on the provided specifications. This includes setting up service parameters to align with organizational requirements and performance standards.
  
- **Environment Isolation**: Manages isolated environments within GCP, ensuring resources are organized and segregated according to organizational needs, which aids in maintaining clear boundaries and reducing resource conflicts.
  
- **Exported Stack Outputs**: Captures essential outputs such as service URLs and resource identifiers in `status.outputs`. These outputs facilitate integration with other infrastructure components, enabling effective monitoring, management, and automation workflows.
  
- **Scalability and Flexibility**: Designed to accommodate a wide range of Cloud Run configurations, the module supports varying levels of complexity and can be easily extended to meet evolving infrastructure demands, ensuring long-term adaptability.
  
- **Error Handling**: Implements robust error handling mechanisms to promptly identify and report issues encountered during deployment or configuration processes, aiding in swift troubleshooting and maintaining infrastructure integrity.

## Installation

To integrate the GCP Cloud Run Pulumi Module into your project, retrieve it from the [GitHub repository](https://github.com/your-repo/gcp-cloud-run-pulumi-module). Ensure that you have both Pulumi and Go installed and properly configured in your development environment.

```shell
git clone https://github.com/your-repo/gcp-cloud-run-pulumi-module.git
cd gcp-cloud-run-pulumi-module
```

## Usage

Refer to the [example section](#examples) for usage instructions.

## Module Details

### Input Configuration

The module expects a `GcpCloudRunStackInput` which includes:

- **Pulumi Input**: Configuration details required by Pulumi for managing the stack, such as stack names, project settings, and any necessary Pulumi configurations.
- **Target API Resource**: The `GcpCloudRun` resource defining the desired Cloud Run configuration, including project ID and credential specifications.
- **GCP Credential**: Specifications for the GCP credentials used to authenticate and authorize Pulumi operations, ensuring secure interactions with GCP services.

### Exported Outputs

Upon successful execution, the module exports the following outputs to `status.outputs`:

- **Service URL**: The URL of the deployed Cloud Run service, facilitating client interactions and service accessibility.
- **Resource ID**: The unique identifier assigned to the created Cloud Run resource, which can be used for management and monitoring purposes.

These outputs facilitate integration with other infrastructure components, provide essential information for monitoring and management, and enable automation workflows to utilize the deployed Cloud Run resources effectively.

## Contributing

We welcome contributions to enhance the GCP Cloud Run Pulumi Module. Please refer to our [contribution guidelines](CONTRIBUTING.md) for more information on how to get involved, including coding standards, submission processes, and best practices.

## License

This project is licensed under the [MIT License](LICENSE), granting you the freedom to use, modify, and distribute the software with minimal restrictions. Please review the LICENSE file for more details.

## Support

For support, please contact our [support team](mailto:support@planton.cloud). We are here to help you with any issues, questions, or feedback you may have regarding the GCP Cloud Run Pulumi Module.

## Acknowledgements

Special thanks to all contributors and the Planton community for their ongoing support and feedback. Your efforts and dedication are instrumental in making this module robust and effective.

## Changelog

A detailed changelog is available in the [CHANGELOG.md](CHANGELOG.md) file. It documents all significant changes, enhancements, bug fixes, and updates made to the GCP Cloud Run Pulumi Module over time.

## Roadmap

We are continuously working to enhance the GCP Cloud Run Pulumi Module. Upcoming features include:

- **Advanced IAM Configurations**: Implementing more granular permission controls for GCP resources to enhance security and compliance.
- **Enhanced Monitoring Integrations**: Integrating with monitoring and logging tools such as Prometheus, Grafana, and the ELK stack for better observability and performance tracking.
- **Support for Additional Cloud Providers**: Extending support to more cloud platforms to increase flexibility and accommodate diverse infrastructure requirements.
- **Automated Scaling and Optimization**: Introducing automated scaling capabilities based on workload demands and performance metrics to optimize resource utilization.
- **Comprehensive Documentation and Tutorials**: Expanding documentation and providing step-by-step tutorials to assist users in effectively leveraging the module's capabilities.

Stay tuned for more updates as we continue to develop and refine the module to meet your infrastructure management needs.

## Contact

For any inquiries or feedback, please reach out to us at [contact@planton.cloud](mailto:contact@planton.cloud). We value your input and are committed to providing the support you need to effectively manage your GCP Cloud Run resources.

## Disclaimer

*This project is maintained by Planton and is not affiliated with any third-party services unless explicitly stated. While we strive to ensure the accuracy and reliability of this module, users are encouraged to review and test configurations in their environments.*

## Security

If you discover any security vulnerabilities, please report them responsibly by contacting our security team at [security@planton.cloud](mailto:security@planton.cloud). We take security seriously and are committed to addressing any issues promptly to protect our users and their infrastructure.

## Code of Conduct

Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) when interacting with the project. We are committed to fostering an inclusive and respectful community where all contributors feel welcome and valued.

## References

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [GCP Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Planton APIs](https://buf.build/openmcf/apis/docs)

## Getting Started

To get started with the GCP Cloud Run Pulumi Module, follow the installation instructions above and refer to the upcoming examples section for detailed usage guidelines. Our comprehensive documentation will guide you through configuring your API resources, setting up Pulumi stacks, and deploying your Cloud Run configurations with ease.
