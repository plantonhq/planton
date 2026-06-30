# Snowflake Database Pulumi Module

## Key Features

### API Resource Features

- **Standardized Structure**: The `SnowflakeDatabase` API resource adheres to a consistent schema with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields. This uniformity ensures compatibility and ease of integration within Kubernetes-like environments, promoting seamless workflow incorporation and tooling interoperability.
  
- **Configurable Specifications**:
  - **Snowflake Credential ID**: Securely reference the Snowflake credentials required to set up the Pulumi provider, ensuring authenticated and authorized interactions with Snowflake services.
  - **Catalog Configuration**: Define the default catalog for Iceberg tables, enabling efficient data organization and management.
  - **Data Retention Policies**: Specify the number of days for Time Travel actions and data retention, ensuring compliance with organizational data governance policies.
  - **Collation and Serialization**: Configure default collation settings and storage serialization policies to optimize database performance and interoperability.
  - **Security Settings**: Manage security-related configurations such as IAM bindings, public schema settings, and trace levels to maintain robust data protection and compliance.

- **Validation and Compliance**: Incorporates stringent validation rules to ensure all configurations adhere to established standards and best practices, minimizing the risk of misconfigurations and enhancing overall system reliability.

### Pulumi Module Features

- **Automated Snowflake Provider Setup**: Leverages the provided Snowflake credentials to automatically configure the Pulumi Snowflake provider, enabling seamless and secure interactions with Snowflake resources.
  
- **Database Management**: Streamlines the creation and management of Snowflake databases based on the provided specifications. This includes setting up database parameters such as catalog, data retention, and security settings to align with organizational requirements.
  
- **Environment Isolation**: Manages isolated environments within Snowflake, ensuring resources are organized and segregated according to organizational needs, which aids in maintaining clear boundaries and reducing resource conflicts.
  
- **Exported Stack Outputs**: Captures essential outputs such as the database ID, bootstrap endpoint, Confluent Resource Name (CRN), and REST endpoint in `status.outputs`. These outputs facilitate integration with other infrastructure components, enabling effective monitoring, management, and automation workflows.
  
- **Scalability and Flexibility**: Designed to accommodate a wide range of Snowflake database configurations, the module supports varying levels of complexity and can be easily extended to meet evolving infrastructure demands, ensuring long-term adaptability.
  
- **Error Handling**: Implements robust error handling mechanisms to promptly identify and report issues encountered during deployment or configuration processes, aiding in swift troubleshooting and maintaining infrastructure integrity.

## Installation

To integrate the Snowflake Database Pulumi Module into your project, retrieve it from the [GitHub repository](https://github.com/your-repo/snowflake-database-pulumi-module). Ensure that you have both Pulumi and Go installed and properly configured in your development environment.

```shell
git clone https://github.com/your-repo/snowflake-database-pulumi-module.git
cd snowflake-database-pulumi-module
```

## Usage

Refer to the [example section](#examples) for usage instructions.

## Module Details

### Input Configuration

The module expects a `SnowflakeDatabaseStackInput` which includes:

- **Pulumi Input**: Configuration details required by Pulumi for managing the stack, such as stack names, project settings, and any necessary Pulumi configurations.
- **Target API Resource**: The `SnowflakeDatabase` resource defining the desired database configuration, including catalog settings, data retention policies, and security configurations.
- **Snowflake Credential**: Specifications for the Snowflake credentials used to authenticate and authorize Pulumi operations, ensuring secure interactions with Snowflake services.

### Exported Outputs

Upon successful execution, the module exports the following outputs to `status.outputs`:

- **Database ID**: The unique identifier assigned to the created Snowflake database, which can be used for management and monitoring purposes.
- **Bootstrap Endpoint**: The endpoint used by Snowflake clients to connect to the database (e.g., `https://pkc-00000.us-central1.gcp.snowflake.cloud:443`), facilitating client interactions.
- **Confluent Resource Name (CRN)**: The Confluent Resource Name of the Snowflake database (e.g., `crn://snowflake.cloud/organization=1111aaaa-11aa-11aa-11aa-111111aaaaaa/environment=env-abc123/cloud-database=ldb-abc123`), serving as a unique identifier within Snowflake Cloud.
- **REST Endpoint**: The REST endpoint of the Snowflake database (e.g., `https://pkc-00000.us-central1.gcp.snowflake.cloud:443`), enabling RESTful interactions with the database for management and monitoring.

These outputs facilitate integration with other infrastructure components, provide essential information for monitoring and management, and enable automation workflows to utilize the deployed Snowflake resources effectively.

## Contributing

We welcome contributions to enhance the Snowflake Database Pulumi Module. Please refer to our [contribution guidelines](CONTRIBUTING.md) for more information on how to get involved, including coding standards, submission processes, and best practices.

## License

This project is licensed under the [MIT License](LICENSE), granting you the freedom to use, modify, and distribute the software with minimal restrictions. Please review the LICENSE file for more details.

## Support

For support, please contact our [support team](mailto:support@planton.ai). We are here to help you with any issues, questions, or feedback you may have regarding the Snowflake Database Pulumi Module.

## Acknowledgements

Special thanks to all contributors and the Planton community for their ongoing support and feedback. Your efforts and dedication are instrumental in making this module robust and effective.

## Changelog

A detailed changelog is available in the [CHANGELOG.md](CHANGELOG.md) file. It documents all significant changes, enhancements, bug fixes, and updates made to the Snowflake Database Pulumi Module over time.

## Roadmap

We are continuously working to enhance the Snowflake Database Pulumi Module. Upcoming features include:

- **Advanced IAM Configurations**: Implementing more granular permission controls for Snowflake resources to enhance security and compliance.
- **Enhanced Monitoring Integrations**: Integrating with monitoring and logging tools such as Prometheus, Grafana, and the ELK stack for better observability and performance tracking.
- **Support for Additional Cloud Providers**: Extending support to more cloud platforms to increase flexibility and accommodate diverse infrastructure requirements.
- **Automated Scaling and Optimization**: Introducing automated scaling capabilities based on workload demands and performance metrics to optimize resource utilization.
- **Comprehensive Documentation and Tutorials**: Expanding documentation and providing step-by-step tutorials to assist users in effectively leveraging the module's capabilities.

Stay tuned for more updates as we continue to develop and refine the module to meet your infrastructure management needs.

## Contact

For any inquiries or feedback, please reach out to us at [contact@planton.ai](mailto:contact@planton.ai). We value your input and are committed to providing the support you need to effectively manage your Snowflake databases.

## Disclaimer

*This project is maintained by Planton and is not affiliated with any third-party services unless explicitly stated. While we strive to ensure the accuracy and reliability of this module, users are encouraged to review and test configurations in their environments.*

## Security

If you discover any security vulnerabilities, please report them responsibly by contacting our security team at [security@planton.ai](mailto:security@planton.ai). We take security seriously and are committed to addressing any issues promptly to protect our users and their infrastructure.

## Code of Conduct

Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) when interacting with the project. We are committed to fostering an inclusive and respectful community where all contributors feel welcome and valued.

## References

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Snowflake Documentation](https://docs.snowflake.com/)
- [Planton APIs](https://buf.build/planton/apis/docs)

## Getting Started

To get started with the Snowflake Database Pulumi Module, follow the installation instructions above and refer to the upcoming examples section for detailed usage guidelines. Our comprehensive documentation will guide you through configuring your API resources, setting up Pulumi stacks, and deploying your Snowflake databases with ease.
