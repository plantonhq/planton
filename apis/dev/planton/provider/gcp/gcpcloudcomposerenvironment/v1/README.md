# GcpCloudComposerEnvironment

Planton component for provisioning Google Cloud Composer environments — managed Apache Airflow services for authoring, scheduling, and monitoring data pipelines.

## Overview

Cloud Composer is Google Cloud's fully managed Apache Airflow service. It provisions and manages the underlying GKE cluster, Cloud SQL metadata database, Airflow web server, and Cloud Storage DAG bucket, allowing you to focus on writing and deploying DAGs rather than managing infrastructure.

This component targets **Composer 2.x and 3** (Composer 1.x is excluded as deprecated by Google). It supports both VPC peering and Private Service Connect (PSC) networking models, flexible workload sizing, CMEK encryption, maintenance windows, and recovery snapshots.

## Key Features

- **Managed Airflow** — Fully managed Apache Airflow with automatic updates and patching
- **Networking options** — VPC peering (Composer 2.x) or Private Service Connect (Composer 2.x and 3)
- **Workload sizing** — Configure CPU, memory, and storage for scheduler, workers, web server, triggerer, and DAG processor
- **Software configuration** — Specify Composer/Airflow image versions, PyPI packages, and Airflow config overrides
- **Private networking** — Private IP environments with optional private endpoints
- **CMEK encryption** — Customer-managed encryption keys for all Composer-managed resources
- **Maintenance windows** — Schedule maintenance operations to minimize disruption
- **Recovery snapshots** — Scheduled snapshots for disaster recovery
- **Access control** — IP-based allowlisting for the Airflow web server UI

## Supported Composer Versions

This component supports:
- **Composer 2.x** — VPC peering or Private Service Connect networking
- **Composer 3** — Private Service Connect networking (default), private environments, DAG processor

**Composer 1.x is not supported** — Google has deprecated Composer 1.x, and this component excludes all Composer 1.x-specific fields.

## Key Configuration Areas

### Networking

**Composer 2.x:**
- **VPC peering** — Traditional networking model using `nodeConfig.network` and `nodeConfig.subnetwork`
- **Private Service Connect** — Modern networking model using `privateEnvironmentConfig.connectionType: PRIVATE_SERVICE_CONNECT`

**Composer 3:**
- **Private Service Connect** — Default networking using `nodeConfig.composerNetworkAttachment`
- **Private environments** — Enable with `enablePrivateEnvironment: true` for no public IP endpoints

### Software Configuration

- **Image version** — Specify Composer and Airflow versions (e.g., `composer-2.9.7-airflow-2.9.3`)
- **PyPI packages** — Install custom Python packages required by your DAGs
- **Airflow config overrides** — Override Airflow configuration properties (e.g., `core-dags_are_paused_at_creation: "True"`)
- **Environment variables** — Set custom environment variables for Airflow components

### Workloads Configuration

Configure resource allocation for each Airflow component:

- **Scheduler** — Parses DAGs, manages task scheduling, triggers task instances
- **Workers** — Execute tasks defined in DAGs (supports autoscaling with min/max counts)
- **Web server** — Airflow UI for monitoring and managing DAGs
- **Triggerer** — Monitors deferred tasks and resumes them when conditions are met (critical for deferrable operators)
- **DAG processor** — Parses DAG files independently (Composer 3 only)

### Private Networking

**Composer 2.x:**
- Configure via `privateEnvironmentConfig` with `enablePrivateEndpoint: true`
- Choose connection type: `VPC_PEERING` or `PRIVATE_SERVICE_CONNECT`
- Specify IP ranges for GKE master, Cloud SQL, and Composer internal components

**Composer 3:**
- Enable private environment with `enablePrivateEnvironment: true`
- Use `nodeConfig.composerNetworkAttachment` for PSC networking
- Optionally enable `enablePrivateBuildsOnly: true` to restrict package builds to private connectivity

### CMEK Encryption

Encrypt all Composer-managed resources (GKE nodes, Cloud SQL, Cloud Storage) with a customer-managed encryption key via `kmsKeyName`. The KMS key must be in the same region as the Composer environment.

### Maintenance Windows

Define when GCP may perform scheduled maintenance via `maintenanceWindow`:
- Start and end times in RFC3339 format
- Recurrence pattern in RFC5545 RRULE format (e.g., `FREQ=WEEKLY;BYDAY=TU,WE,TH`)
- Minimum window duration: 12 hours

### Recovery Configuration

Enable scheduled snapshots for disaster recovery via `recoveryConfig`:
- Cloud Storage location for snapshots
- Cron schedule for snapshot creation (Unix-cron format)
- Time zone for the schedule

### Access Control

Restrict web server access to specific IP ranges via `webServerNetworkAccessControl`:
- Define allowed IP ranges (CIDR notation)
- Optional descriptions for each range

## Key Fields

| Field | Required | Description |
|-------|----------|-------------|
| `projectId` | Yes | GCP project ID |
| `region` | Yes | GCP region (e.g., `us-central1`) |
| `environmentName` | No | Composer environment name; defaults to `metadata.name` |
| `environmentSize` | No | `ENVIRONMENT_SIZE_SMALL`, `ENVIRONMENT_SIZE_MEDIUM`, or `ENVIRONMENT_SIZE_LARGE` |
| `nodeConfig` | No | Networking and compute settings for GKE nodes |
| `softwareConfig` | No | Airflow software configuration (image version, packages, config overrides) |
| `privateEnvironmentConfig` | No | Private networking for Composer 2.x |
| `workloadsConfig` | No | Resource allocation for Airflow components |
| `kmsKeyName` | No | Customer-managed encryption key |
| `maintenanceWindow` | No | Scheduled maintenance window |
| `recoveryConfig` | No | Disaster recovery snapshots |
| `webServerNetworkAccessControl` | No | IP-based access restrictions for web server |

## Environment Sizes

| Size | Description | Use Case |
|------|-------------|----------|
| `ENVIRONMENT_SIZE_SMALL` | Minimal capacity | Development, testing |
| `ENVIRONMENT_SIZE_MEDIUM` | Standard capacity | Small to medium production workloads |
| `ENVIRONMENT_SIZE_LARGE` | High capacity | Large-scale production workloads |

## Resilience Modes

| Mode | Description | Availability |
|------|-------------|-------------|
| `STANDARD_RESILIENCE` | Single-zone deployment | Standard availability |
| `HIGH_RESILIENCE` | Multi-zone redundancy | Increased availability (Composer 2.1.15+) |

## Examples

See [examples.md](examples.md) for copy-paste ready YAML manifests.

## Further Reading

- [Research & design documentation](docs/README.md) — Architecture, networking models, workload sizing, security, best practices
