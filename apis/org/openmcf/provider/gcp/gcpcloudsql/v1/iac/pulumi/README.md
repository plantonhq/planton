# GCP Cloud SQL Pulumi Module

## Overview

This Pulumi module provides automated deployment and management of Google Cloud SQL instances for MySQL and PostgreSQL databases. It handles instance provisioning, network configuration, high availability setup, backup configuration, and database flag management.

## Key Features

### API Resource Features

- **Standardized Structure**: The `GcpCloudSql` API resource adheres to a consistent schema with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields, ensuring compatibility and ease of integration within Kubernetes-like environments.
  
- **Configurable Specifications**:
  - **GCP Project ID**: Specifies the GCP project where the Cloud SQL instance will be deployed.
  - **Database Engine**: Support for both MySQL and PostgreSQL with configurable versions.
  - **Instance Tier**: Flexible machine type selection from shared-core to high-memory configurations.
  - **Storage Configuration**: Configurable SSD storage from 10GB to 65TB.
  - **Network Settings**: Private IP via VPC peering and/or public IP with authorized network restrictions.
  - **High Availability**: Optional regional HA with automatic failover capabilities.
  - **Automated Backups**: Configurable backup schedules with customizable retention periods.
  - **Database Flags**: Custom database configuration flags for fine-tuning.

- **Validation and Compliance**: Incorporates stringent validation rules to ensure all configurations adhere to best practices and GCP requirements.

### Pulumi Module Features

- **Automated GCP Provider Setup**: Leverages the provided GCP credentials to automatically configure the Pulumi GCP provider.
  
- **Cloud SQL Instance Management**: Streamlines the creation and management of Cloud SQL database instances based on the provided specifications.
  
- **Network Configuration**: Handles VPC peering for private connectivity and authorized network configuration for public access.
  
- **High Availability Setup**: Configures regional high availability with automatic failover when enabled.
  
- **Backup Configuration**: Sets up automated daily backups with point-in-time recovery capabilities.
  
- **Exported Stack Outputs**: Captures essential outputs such as instance name, connection name, IP addresses, and self-link in `status.outputs`.
  
- **Error Handling**: Implements robust error handling mechanisms to identify and report issues during deployment.

## Installation

To integrate the GCP Cloud SQL Pulumi Module into your project, retrieve it from the GitHub repository. Ensure that you have both Pulumi and Go installed and properly configured.

```shell
git clone https://github.com/plantonhq/openmcf.git
cd openmcf/apis/project/planton/provider/gcp/gcpcloudsql/v1/iac/pulumi
```

## Usage

Refer to the [example section](examples.md) for detailed usage instructions.

## Module Details

### Input Configuration

The module expects a `GcpCloudSqlStackInput` which includes:

- **Target API Resource**: The `GcpCloudSql` resource defining the desired database configuration.
- **GCP Credential**: Specifications for the GCP credentials used to authenticate and authorize Pulumi operations.

### Exported Outputs

Upon successful execution, the module exports the following outputs to `status.outputs`:

- **instance_name**: Name of the Cloud SQL instance
- **connection_name**: Full connection name in the format `project:region:instance`
- **private_ip**: Private IP address (if private IP is enabled)
- **public_ip**: Public IP address
- **self_link**: GCP resource self link for the Cloud SQL instance

These outputs facilitate integration with other infrastructure components and enable automation workflows.

## Configuration Examples

### Basic MySQL Instance

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: mysql-db
spec:
  projectId: my-gcp-project
  region: us-central1
  database_engine: MYSQL
  database_version: MYSQL_8_0
  tier: db-n1-standard-1
  storage_gb: 10
  root_password: SecurePassword123!
```

### PostgreSQL with High Availability

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudSql
metadata:
  name: postgres-ha
spec:
  projectId: my-gcp-project
  region: us-central1
  database_engine: POSTGRESQL
  database_version: POSTGRES_15
  tier: db-n1-standard-2
  storage_gb: 50
  high_availability:
    enabled: true
    zone: us-central1-b
  backup:
    enabled: true
    start_time: "03:00"
    retention_days: 7
  root_password: SecurePassword123!
```

## Network Configuration

### Private IP

To enable private IP connectivity:

```yaml
network:
  vpc_id: projects/my-project/global/networks/my-vpc
  private_ip_enabled: true
```

Note: VPC peering with the Cloud SQL service must be configured in advance.

### Public IP with Authorized Networks

To restrict public IP access to specific CIDR ranges:

```yaml
network:
  authorized_networks:
    - 203.0.113.0/24
    - 198.51.100.0/24
```

## Database Flags

Custom database configuration flags can be specified:

```yaml
database_flags:
  max_connections: "200"
  slow_query_log: "on"
```

Refer to MySQL or PostgreSQL documentation for available flags.

## Deployment

### Using Pulumi CLI

```shell
pulumi up --stack <org>/<stack-name>/<environment>
```

### Using OpenMCF CLI

```shell
openmcf pulumi up --manifest gcpcloudsql.yaml --stack <org>/<stack-name>/<environment>
```

## Contributing

We welcome contributions to enhance the GCP Cloud SQL Pulumi Module. Please refer to our contribution guidelines for more information.

## License

This project is licensed under the MIT License. Please review the LICENSE file for more details.

## Support

For support, please contact our support team at support@planton.ai.

## References

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [GCP Cloud SQL Documentation](https://cloud.google.com/sql/docs)
- [Planton APIs](https://buf.build/openmcf/apis/docs)

