# Confluent Cloud Kafka Pulumi Module

## Overview

The **Confluent Cloud Kafka Pulumi Module** is designed to simplify the deployment and management of Confluent Cloud Kafka clusters across multiple cloud providers. By leveraging Google Cloud Platform (GCP), AWS, and Azure, this module integrates seamlessly with Planton's unified API framework, which models every API resource using a Kubernetes-like structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields. The `ConfluentKafka` resource defines the necessary specifications for provisioning Kafka clusters in Confluent Cloud, enabling developers to manage their messaging infrastructure as code effortlessly.

By utilizing this Pulumi module, developers can automate the creation of Kafka clusters with specified configurations such as cloud provider selection, availability zones, and environment isolation. The module interacts with Confluent credentials and other necessary specifications provided in the resource definition, ensuring a streamlined and consistent deployment process. Additionally, the outputs from the deployment, including cluster IDs and endpoints, are captured in the resource's `status.outputs`, allowing users to monitor and manage their Kafka infrastructure directly through the `ConfluentKafka` resource.

## Important Note

*This module is not completely implemented. Certain features may be missing or not fully functional. Future updates will address these limitations.*

## Key Features

### API Resource Features

- **Standardized Structure**: The `ConfluentKafka` API resource follows a consistent schema with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields. This uniformity ensures compatibility and ease of integration within Kubernetes-like environments, promoting seamless workflow incorporation and tooling interoperability.

- **Configurable Specifications**:
  - **Confluent Credential ID**: Securely reference the Confluent credentials required to set up the Pulumi provider, ensuring authenticated and authorized interactions with Confluent Cloud services.
  - **Cloud Provider Selection**: Specify the target cloud provider (`AWS`, `AZURE`, `GCP`) for deploying the Kafka cluster, enabling flexibility and support for multi-cloud deployments.
  - **Availability Configuration**: Define the availability level of the Kafka cluster (`SINGLE_ZONE`, `MULTI_ZONE`, `LOW`, `HIGH`) to meet varying resilience and performance requirements.
  - **Environment Isolation**: Utilize the `environment` field to represent an isolated namespace for Confluent resources, facilitating organizational management and resource segregation.

- **Validation and Compliance**: Incorporates stringent validation rules to ensure all configurations adhere to established standards and best practices, minimizing the risk of misconfigurations and enhancing overall system reliability.

### Pulumi Module Features

- **Automated Confluent Provider Setup**: Leverages the provided Confluent credentials to automatically configure the Pulumi Confluent provider, enabling seamless and secure interactions with Confluent Cloud resources.

- **Kafka Cluster Management**: Streamlines the creation and management of Confluent Cloud Kafka clusters based on the provided specifications, including cloud provider selection and availability settings to align with organizational requirements.

- **Environment Isolation**: Manages isolated environments within Confluent Cloud, ensuring resources are organized and segregated according to organizational needs, which aids in maintaining clear boundaries and reducing resource conflicts.

- **Exported Stack Outputs**: Captures essential outputs such as the cluster ID, bootstrap endpoint, Confluent Resource Name (CRN), and REST endpoint in `status.outputs`. These outputs facilitate integration with other infrastructure components, enabling effective monitoring, management, and automation workflows.

- **Scalability and Flexibility**: Designed to accommodate a wide range of Kafka cluster configurations, the module supports varying levels of complexity and can be easily extended to meet evolving infrastructure demands, ensuring long-term adaptability.

- **Error Handling**: Implements robust error handling mechanisms to promptly identify and report issues encountered during deployment or configuration processes, aiding in swift troubleshooting and maintaining infrastructure integrity.

## Installation

To integrate the Confluent Cloud Kafka Pulumi Module into your project, retrieve it from the [GitHub repository](https://github.com/your-repo/confluent-kafka-pulumi-module). Ensure that you have both Pulumi and Go installed and properly configured in your development environment.

```shell
git clone https://github.com/your-repo/confluent-kafka-pulumi-module.git
cd confluent-kafka-pulumi-module
```

## Usage

Refer to the [example section](#examples) for usage instructions.

## Module Details

### Input Configuration

The module expects a `ConfluentKafkaStackInput` which includes:

- **Pulumi Input**: Configuration details required by Pulumi for managing the stack, such as stack names, project settings, and any necessary Pulumi configurations.
- **Target API Resource**: The `ConfluentKafka` resource defining the desired Kafka cluster configuration, including cloud provider, availability, and environment specifications.
- **Confluent Credential**: Specifications for the Confluent credentials used to authenticate and authorize Pulumi operations, ensuring secure interactions with Confluent Cloud.

### Exported Outputs

Upon successful execution, the module exports the following outputs to `status.outputs`:

- **Cluster ID**: The unique identifier assigned to the created Kafka cluster, which can be used for management and monitoring purposes.
- **Bootstrap Endpoint**: The endpoint used by Kafka clients to connect to the Kafka cluster (e.g., `SASL_SSL://pkc-00000.us-central1.gcp.confluent.cloud:9092`), facilitating client interactions.
- **Confluent Resource Name (CRN)**: The Confluent Resource Name of the Kafka cluster (e.g., `crn://confluent.cloud/organization=1111aaaa-11aa-11aa-11aa-111111aaaaaa/environment=env-abc123/cloud-cluster=lkc-abc123`), serving as a unique identifier within Confluent Cloud.
- **REST Endpoint**: The REST endpoint of the Kafka cluster (e.g., `https://pkc-00000.us-central1.gcp.confluent.cloud:443`), enabling RESTful interactions with the cluster for management and monitoring.

These outputs facilitate integration with other infrastructure components, provide essential information for monitoring and management, and enable automation workflows to utilize the deployed Kafka resources effectively.

## Contributing

We welcome contributions to enhance the Confluent Cloud Kafka Pulumi Module. Please refer to our [contribution guidelines](CONTRIBUTING.md) for more information on how to get involved, including coding standards, submission processes, and best practices.

## License

This project is licensed under the [MIT License](LICENSE), granting you the freedom to use, modify, and distribute the software with minimal restrictions. Please review the LICENSE file for more details.

## Support

For support, please contact our [support team](mailto:support@planton.cloud). We are here to help you with any issues, questions, or feedback you may have regarding the Confluent Cloud Kafka Pulumi Module.

## Acknowledgements

Special thanks to all contributors and the Planton community for their ongoing support and feedback. Your efforts and dedication are instrumental in making this module robust and effective.

## Changelog

A detailed changelog is available in the [CHANGELOG.md](CHANGELOG.md) file. It documents all significant changes, enhancements, bug fixes, and updates made to the Confluent Cloud Kafka Pulumi Module over time.

## Roadmap

We are continuously working to enhance the Confluent Cloud Kafka Pulumi Module. Upcoming features include:

- **Advanced IAM Configurations**: Implementing more granular permission controls for Kafka resources to enhance security and compliance.
- **Enhanced Monitoring Integrations**: Integrating with monitoring and logging tools such as Prometheus, Grafana, and the ELK stack for better observability and performance tracking.
- **Support for Additional Cloud Providers**: Extending support to more cloud platforms to increase flexibility and accommodate diverse infrastructure requirements.
- **Automated Scaling and Optimization**: Introducing automated scaling capabilities based on workload demands and performance metrics to optimize resource utilization.
- **Comprehensive Documentation and Tutorials**: Expanding documentation and providing step-by-step tutorials to assist users in effectively leveraging the module's capabilities.

Stay tuned for more updates as we continue to develop and refine the module to meet your infrastructure management needs.

## Contact

For any inquiries or feedback, please reach out to us at [contact@planton.cloud](mailto:contact@planton.cloud). We value your input and are committed to providing the support you need to effectively manage your Confluent Cloud Kafka resources.

## Disclaimer

*This project is maintained by Planton and is not affiliated with any third-party services unless explicitly stated. While we strive to ensure the accuracy and reliability of this module, users are encouraged to review and test configurations in their environments.*

## Security

If you discover any security vulnerabilities, please report them responsibly by contacting our security team at [security@planton.cloud](mailto:security@planton.cloud). We take security seriously and are committed to addressing any issues promptly to protect our users and their infrastructure.

## Code of Conduct

Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) when interacting with the project. We are committed to fostering an inclusive and respectful community where all contributors feel welcome and valued.

## References

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Confluent Cloud Documentation](https://docs.confluent.io/cloud/)
- [Planton APIs](https://buf.build/openmcf/apis/docs)

## Getting Started

To get started with the Confluent Cloud Kafka Pulumi Module, follow the installation instructions above and refer to the upcoming examples section for detailed usage guidelines. Our comprehensive documentation will guide you through configuring your API resources, setting up Pulumi stacks, and deploying your Kafka clusters with ease.
