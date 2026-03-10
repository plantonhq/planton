# Overview

The **Postgres Kubernetes API resource** is designed to simplify the deployment and management of PostgreSQL databases within Kubernetes environments. This API resource allows users to configure PostgreSQL containers, resource allocations, and ingress settings efficiently, ensuring consistency and reliability in managing PostgreSQL instances.

## Why We Created This API Resource

Deploying and managing PostgreSQL databases in Kubernetes can be complex due to the need to handle container resources, replicas, and ingress configurations. This API resource was developed to:

- **Simplify PostgreSQL Deployment**: Provides an easy-to-use interface for deploying PostgreSQL in Kubernetes environments.
- **Ensure Consistency**: Offers a standardized approach to deploying PostgreSQL across different Kubernetes clusters and environments.
- **Optimize Resource Management**: Allows fine-tuning of CPU, memory, and storage allocations for PostgreSQL containers.
- **Streamline Ingress Management**: Facilitates ingress configuration to allow secure access to PostgreSQL instances.

## Key Features

### Environment Integration

- **Environment Info**: Automatically integrates with the Planton environment management system, ensuring that PostgreSQL is deployed in the right context.
- **Stack Job Settings**: Supports stack-update settings for consistent infrastructure-as-code deployments.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: Specifies the Kubernetes credentials (`kubernetes_credential_id`) for securely deploying and managing the PostgreSQL container in Kubernetes.

### PostgreSQL Container Configuration

#### Resource Management

- **Replicas**: Define the number of PostgreSQL replicas to ensure high availability and fault tolerance. The recommended default is 1 replica.

- **Container Resources**: Customize the CPU and memory resources for the PostgreSQL container to ensure optimal performance. The recommended default values are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`

- **Disk Size**: Configure the storage size for each PostgreSQL instance. The default value is `1Gi`, but you can specify a different size based on your requirements.

### Namespace Management

- **Flexible Namespace Control**: Choose whether the component creates the namespace or uses an existing one
  - Set `create_namespace: true` to have the component create and manage the namespace
  - Set `create_namespace: false` to deploy into an existing namespace that you manage separately
- **Use Cases**:
  - **New deployments**: Use `create_namespace: true` for component-managed namespace with proper labels
  - **Shared namespaces**: Use `create_namespace: false` when multiple components share a namespace
  - **Pre-configured environments**: Use `create_namespace: false` when namespaces are created by platform teams

### Ingress Configuration

- **Ingress Spec**: Manage and control external access to the PostgreSQL instance by configuring ingress with a custom hostname. When enabled, creates a LoadBalancer service with external-dns annotations for automatic DNS configuration. Users specify the exact hostname (e.g., `postgres.example.com`) instead of auto-constructed patterns, providing full control over the ingress endpoint.

### User Configuration

- **Declarative User Management**: Define PostgreSQL users/roles via the `users` field. Each user has a `name` and optional `flags` for permissions.
- **User Flags**: Common flags include `createdb`, `superuser`, `createrole`, `inherit`, `login`, and `replication`. Empty flags (`[]`) creates a standard user with login privileges.
- **Example Configuration**:
  ```yaml
  users:
    - name: app_user
      flags: []           # Standard user
    - name: analytics
      flags:
        - createdb        # Can create databases
  ```
- **Required for Database Owners**: Users must be declared before being referenced as database owners.

### Database Configuration

- **Multiple Databases**: Specify databases to create during cluster initialization via the `databases` field. This is a map where keys are database names and values are owner role names.
- **Owner Role Requirement**: Owner roles must be declared in the `users` field before being referenced as database owners. The operator will skip database creation if the owner doesn't exist.
- **Example Configuration**:
  ```yaml
  users:
    - name: app_user
      flags: []
  databases:
    app_database: app_user
  ```
- **Default Behavior**: If no databases are specified, only the default `postgres` database is available.

### Backup Configuration

- **Automatic Backups to R2/S3**: Configure automated continuous backups to Cloudflare R2 or any S3-compatible storage using WAL-G. Set custom backup schedules and S3 prefixes per database.
- **Operator-Level Defaults**: Inherits backup configuration from the PostgreSQL operator if not specified, allowing for centralized backup management.
- **Per-Database Overrides**: Override operator-level backup settings with custom schedules, S3 paths, or enable/disable backups on a per-database basis.

### Disaster Recovery

- **Cross-Cluster Restore**: Restore PostgreSQL databases from R2/S3 backups to any Kubernetes cluster, independent of source cluster availability. Enables true disaster recovery scenarios.
- **Technology-Agnostic Design**: Restore API works with multiple PostgreSQL operators (Zalando, Percona, CloudNativePG) through a unified interface.
- **Two-Stage Workflow**:
  - **Stage 1 (Bootstrap)**: Set `restore.enabled: true` to create a read-only standby database restored from backup, allowing data validation before promotion.
  - **Stage 2 (Promote)**: Set `restore.enabled: false` to promote the standby to a read-write primary database.
- **Component Independence**: Each database can specify its own R2/S3 credentials for complete deployment independence.
- **Graceful Fallback**: Bucket name and credentials fall back to operator-level configuration if not specified per-database.

## Benefits

- **Simplified Deployment**: Reduces the complexity of deploying PostgreSQL on Kubernetes by providing an easy-to-use API resource.
- **Consistent Configuration**: Ensures PostgreSQL deployments follow a consistent configuration across different Kubernetes clusters and environments.
- **Resource Optimization**: Enables fine-tuning of CPU, memory, and storage allocations to optimize the performance of PostgreSQL instances.
- **Secure Access**: Configures ingress rules to allow secure and controlled access to PostgreSQL instances.
- **Flexible Namespace Management**: Supports both component-managed and externally-managed namespaces for different deployment scenarios.
- **Automated Backups**: Continuous backup to R2/S3 with configurable schedules, ensuring data protection without manual intervention.
- **Disaster Recovery Ready**: Restore databases across clusters from backup files alone, enabling recovery even when source cluster is completely destroyed.
- **Operator Flexibility**: Technology-agnostic restore design works across different PostgreSQL operators without API changes.
