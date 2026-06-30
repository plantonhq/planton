# Overview

The **GCP Cloud SQL API Resource** provides a consistent and standardized interface for deploying and managing relational databases on Google Cloud SQL within our infrastructure. This resource simplifies the process of creating and configuring fully managed MySQL and PostgreSQL database instances on Google Cloud Platform (GCP), allowing users to build scalable data storage solutions without managing underlying database infrastructure.

## Purpose

We developed this API resource to streamline the deployment and management of relational databases using GCP Cloud SQL. By offering a unified interface, it reduces the complexity involved in setting up and configuring production-grade databases, enabling users to:

- **Easily Deploy Cloud SQL Instances**: Quickly create MySQL and PostgreSQL database instances in specified GCP projects.
- **Simplify Configuration**: Abstract the complexities of setting up Cloud SQL, including networking, high availability, and backup configurations.
- **Integrate Seamlessly**: Utilize existing GCP credentials and integrate with VPC networks for private connectivity.
- **Focus on Applications**: Allow developers to concentrate on building applications rather than managing database infrastructure.

## Key Features

- **Consistent Interface**: Aligns with our existing APIs for deploying cloud infrastructure and managed services.
- **Multiple Database Engines**: Support for both MySQL and PostgreSQL with configurable versions.
- **Simplified Deployment**: Automates the provisioning of Cloud SQL instances with necessary configurations.
- **Network Security**: Support for private IP connectivity through VPC networks and authorized network configurations for public access.
- **High Availability**: Optional regional high availability configuration with automatic failover.
- **Automated Backups**: Configurable backup schedules with customizable retention periods.
- **Scalability**: Configure instance tiers and storage to match workload requirements.
- **Flexible Configuration**: Support for database flags to fine-tune database behavior.

## Use Cases

- **Application Databases**: Deploy production-grade relational databases for web and mobile applications.
- **Microservices Data Layer**: Provide dedicated database instances for microservices architectures.
- **Development and Testing**: Quickly spin up database instances for development, staging, and testing environments.
- **Data Migration**: Migrate on-premises or other cloud databases to GCP Cloud SQL.
- **Analytics Workloads**: Set up PostgreSQL instances for analytical queries and reporting.
- **Multi-tenancy**: Deploy separate database instances for different customers or business units.

## Architecture

GCP Cloud SQL instances created through this API resource include:

- **Managed Database Engine**: Fully managed MySQL or PostgreSQL database with automatic updates and maintenance.
- **Network Configuration**: Private IP connectivity through VPC peering and/or public IP with authorized networks.
- **Storage Management**: SSD-backed storage with configurable size (10GB to 65TB).
- **High Availability**: Optional regional HA configuration with synchronous replication to a standby instance.
- **Automated Backups**: Point-in-time recovery with configurable backup windows and retention.
- **Monitoring**: Built-in integration with Google Cloud Monitoring and Logging.

## Configuration Options

### Database Engine

Choose between:
- **MySQL**: Versions like MYSQL_8_0, MYSQL_5_7
- **PostgreSQL**: Versions like POSTGRES_15, POSTGRES_14, POSTGRES_13

### Instance Tier

Select machine types based on workload requirements:
- Shared-core: `db-f1-micro`, `db-g1-small`
- Standard: `db-n1-standard-1`, `db-n1-standard-2`, etc.
- High-memory: `db-n1-highmem-2`, `db-n1-highmem-4`, etc.
- Custom: Custom CPU and memory configurations

### Storage

- Configurable from 10GB to 65,536GB (65TB)
- SSD-backed for high performance
- Automatic storage increase can be enabled

### Networking

- **Private IP**: Connect through VPC peering for secure, low-latency access
- **Public IP**: Access from the internet with authorized network CIDR restrictions
- **Hybrid**: Enable both private and public IP addresses

### High Availability

- Enable regional HA for 99.95% uptime SLA
- Automatic failover to standby instance in different zone
- Synchronous replication for zero data loss

### Backups

- Automated daily backups with configurable start time
- Retention period from 1 to 365 days
- Point-in-time recovery support
- Manual backups can be triggered as needed

## Security

- **Encryption at Rest**: All data automatically encrypted using Google-managed keys
- **Encryption in Transit**: SSL/TLS connections supported
- **IAM Integration**: Database access controlled through GCP IAM
- **Network Isolation**: Private IP connectivity through VPC peering
- **Authorized Networks**: Restrict public IP access to specific CIDR ranges
- **Password Protection**: Required strong root password (minimum 8 characters)

## Future Enhancements

As this resource continues to evolve, future updates will include:

- **Read Replicas**: Support for creating and managing read replicas for scaling read workloads.
- **Database Users**: Automated creation and management of database users with different privileges.
- **Database Creation**: Automatic creation of initial databases within the instance.
- **Enhanced Monitoring**: Custom metrics and alerting configurations.
- **Maintenance Windows**: Configurable maintenance windows for updates.
- **SSL Certificate Management**: Automated SSL certificate provisioning and rotation.
- **Cloud SQL Proxy**: Integration with Cloud SQL proxy for secure connections.
- **Migration Tools**: Built-in support for database migration from other sources.
- **Performance Insights**: Query performance monitoring and optimization recommendations.

