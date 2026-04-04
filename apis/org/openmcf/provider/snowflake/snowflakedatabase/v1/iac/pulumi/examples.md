# Snowflake Database Pulumi Module

## Examples

### Basic Example

This example demonstrates a basic setup of a Snowflake database with minimal configuration.

```yaml
apiVersion: snowflake.openmcf.org/v1
kind: SnowflakeDatabase
metadata:
  name: analytics-db
spec:
  snowflake_credential_id: snowflake-cred-123
  catalog: default_catalog
  comment: "Analytics database for reporting"
  data_retention_time_in_days: 30
  default_ddl_collation: "en_US"
  drop_public_schema_on_creation: false
  enable_console_output: true
  external_volume: "external_vol_1"
  is_transient: false
  log_level: "INFO"
  max_data_extension_time_in_days: 10
  name: analytics_db
  quoted_identifiers_ignore_case: true
  replace_invalid_characters: false
  storage_serialization_policy: "COMPATIBLE"
  suspend_task_after_num_failures: 3
  task_auto_retry_attempts: 2
  trace_level: "OFF"
```

### Example with Advanced Configuration

This example includes advanced configurations such as environment isolation and detailed security settings.

```yaml
apiVersion: snowflake.openmcf.org/v1
kind: SnowflakeDatabase
metadata:
  name: finance-db
spec:
  snowflake_credential_id: snowflake-cred-finance
  catalog: finance_catalog
  comment: "Finance database for transactional data"
  data_retention_time_in_days: 90
  default_ddl_collation: "en_US"
  drop_public_schema_on_creation: true
  enable_console_output: false
  external_volume: "external_vol_finance"
  is_transient: true
  log_level: "DEBUG"
  max_data_extension_time_in_days: 15
  name: finance_db
  quoted_identifiers_ignore_case: false
  replace_invalid_characters: true
  storage_serialization_policy: "OPTIMIZED"
  suspend_task_after_num_failures: 5
  task_auto_retry_attempts: 3
  trace_level: "ALWAYS ON"
```

### Example with Environment Secrets

This example integrates environment secrets managed by Planton's [Snowflake Secrets Manager](https://buf.build/openmcf/apis/docs/main:ai.planton.code2cloud.v1.snowflake.snowflakedatabase).

```yaml
apiVersion: snowflake.openmcf.org/v1
kind: SnowflakeDatabase
metadata:
  name: secure-db
spec:
  snowflake_credential_id: snowflake-cred-secure
  catalog: secure_catalog
  comment: "Secure database with sensitive configurations"
  data_retention_time_in_days: 60
  default_ddl_collation: "en_GB"
  drop_public_schema_on_creation: false
  enable_console_output: true
  external_volume: "external_vol_secure"
  is_transient: false
  log_level: "WARN"
  max_data_extension_time_in_days: 20
  name: secure_db
  quoted_identifiers_ignore_case: true
  replace_invalid_characters: true
  storage_serialization_policy: "COMPATIBLE"
  suspend_task_after_num_failures: 4
  task_auto_retry_attempts: 1
  trace_level: "EVENT OFF"
  env:
    secrets:
      # value before dot 'snowflakesm-secure-env-snowflake-secrets' is the id of the snowflake-secret-manager resource on planton
      # value after dot 'db-password' is one of the secrets list in 'snowflakesm-secure-env-snowflake-secrets' is the id of the snowflake-secret-manager resource on planton
      DB_PASSWORD: ${snowflakesm-secure-env-snowflake-secrets.db-password}
    variables:
      ADMIN_USER: admin
```

### Example with Empty Spec

If the `spec` field is empty, the module is not completely implemented. Below are example configurations, though they may not function as expected.

```yaml
apiVersion: snowflake.openmcf.org/v1
kind: SnowflakeDatabase
metadata:
  name: incomplete-db
spec: {}
```

```yaml
apiVersion: snowflake.openmcf.org/v1
kind: SnowflakeDatabase
metadata:
  name: another-incomplete-db
spec: {}
```

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
- [Planton APIs](https://buf.build/openmcf/apis/docs)
