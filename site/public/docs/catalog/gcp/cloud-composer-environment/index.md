---
title: "Cloud Composer Environment"
description: "Cloud Composer Environment deployment documentation"
icon: "package"
order: 100
componentName: "gcpcloudcomposerenvironment"
---

# GCP Cloud Composer Environment

Deploys a Google Cloud Composer environment — a managed Apache Airflow service that provisions the underlying GKE cluster, Cloud SQL metadata database, Airflow web server, and Cloud Storage DAG bucket. Supports Composer 2.x and 3, with configurable workload sizing, private networking, CMEK encryption, and scheduled recovery snapshots.

## What Gets Created

When you deploy a GcpCloudComposerEnvironment resource, Planton provisions:

- **Cloud Composer Environment** — a `google_composer_environment` resource that manages the full Airflow stack (scheduler, workers, web server, triggerer, metadata database, DAG storage)
- **GKE Cluster** — automatically created and managed by Composer to run Airflow workloads as Kubernetes pods
- **Cloud SQL Instance** — stores Airflow metadata (DAG runs, task instances, variables, connections)
- **Cloud Storage Bucket** — holds DAG files, plugins, and data; the `dag_gcs_prefix` output provides the upload path
- **Airflow Web Server** — hosts the Airflow UI, accessible via the `airflow_uri` output

All infrastructure is managed by Cloud Composer. You configure the environment; Composer handles the lifecycle of the underlying resources.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** with the Cloud Composer API enabled
- **A VPC network and subnetwork** if configuring VPC peering networking (Composer 2.x)
- **A PSC Network Attachment** if using Composer 3 networking
- **A service account** with Composer Worker role (`roles/composer.worker`) if specifying a custom service account
- **A KMS key** if enabling CMEK encryption

## Quick Start

Create a file `composer.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: my-airflow
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpCloudComposerEnvironment.my-airflow
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
```

Deploy:

```shell
planton apply -f composer.yaml
```

This creates a Composer environment with default settings (Composer 2.x, ENVIRONMENT_SIZE_SMALL, public endpoint) in us-central1. The `airflow_uri` output provides the web UI URL.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project for the environment. Can reference a GcpProject resource via `valueFrom`. | Required |
| `region` | `string` | GCP region (e.g., `us-central1`). | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `environmentName` | `string` | `metadata.name` | Explicit GCP resource name. Must be lowercase, letters/numbers/hyphens, 1-64 chars. |
| `environmentSize` | `string` | — | Environment capacity: `ENVIRONMENT_SIZE_SMALL`, `ENVIRONMENT_SIZE_MEDIUM`, `ENVIRONMENT_SIZE_LARGE`. |
| `resilienceMode` | `string` | — | `STANDARD_RESILIENCE` or `HIGH_RESILIENCE` (multi-zone redundancy). Composer 2.1.15+. |
| `kmsKeyName` | `StringValueOrRef` | — | CMEK encryption key for all Composer-managed resources. Can reference GcpKmsKey via `valueFrom`. |
| `enablePrivateEnvironment` | `bool` | `false` | Composer 3 only. Disables the public web server endpoint. |
| `enablePrivateBuildsOnly` | `bool` | `false` | Composer 3 only. Restricts package builds to private connectivity. |
| `nodeConfig.network` | `StringValueOrRef` | — | VPC network for Composer 2.x. Can reference GcpVpc via `valueFrom`. |
| `nodeConfig.subnetwork` | `StringValueOrRef` | — | VPC subnetwork for Composer 2.x. Can reference GcpSubnetwork via `valueFrom`. |
| `nodeConfig.serviceAccount` | `StringValueOrRef` | — | Service account for GKE nodes. Can reference GcpServiceAccount via `valueFrom`. |
| `nodeConfig.tags` | `string[]` | `[]` | Network tags for firewall targeting. |
| `nodeConfig.composerNetworkAttachment` | `string` | — | Composer 3 PSC network attachment. Mutually exclusive with `network`/`subnetwork`. |
| `nodeConfig.composerInternalIpv4CidrBlock` | `string` | — | Composer 3 internal CIDR (/20). |
| `softwareConfig.imageVersion` | `string` | latest | Composer/Airflow version (e.g., `composer-2.9.7-airflow-2.9.3`). |
| `softwareConfig.airflowConfigOverrides` | `map<string,string>` | `{}` | Airflow config overrides (e.g., `core-dags_are_paused_at_creation: "True"`). |
| `softwareConfig.pypiPackages` | `map<string,string>` | `{}` | Custom PyPI packages to install (e.g., `numpy: ">=1.21"`). |
| `softwareConfig.envVariables` | `map<string,string>` | `{}` | Environment variables for Airflow components. |
| `softwareConfig.webServerPluginsMode` | `string` | — | Composer 3 only: `ENABLED` or `DISABLED`. |
| `privateEnvironmentConfig.enablePrivateEndpoint` | `bool` | `false` | Composer 2.x: deny public web server access. |
| `privateEnvironmentConfig.connectionType` | `string` | — | `VPC_PEERING` or `PRIVATE_SERVICE_CONNECT`. |
| `privateEnvironmentConfig.masterIpv4CidrBlock` | `string` | — | CIDR for GKE master (e.g., `172.16.0.0/28`). |
| `privateEnvironmentConfig.cloudSqlIpv4CidrBlock` | `string` | — | CIDR for Cloud SQL. |
| `privateEnvironmentConfig.cloudComposerNetworkIpv4CidrBlock` | `string` | — | CIDR for Composer components. |
| `privateEnvironmentConfig.cloudComposerConnectionSubnetwork` | `string` | — | PSC connection subnetwork. |
| `privateEnvironmentConfig.enablePrivatelyUsedPublicIps` | `bool` | `false` | Allow public IPs from non-RFC1918 ranges. |
| `workloadsConfig.scheduler` | `object` | — | Scheduler: `cpu`, `memoryGb`, `storageGb`, `count`. |
| `workloadsConfig.webServer` | `object` | — | Web server: `cpu`, `memoryGb`, `storageGb`. |
| `workloadsConfig.worker` | `object` | — | Workers: `cpu`, `memoryGb`, `storageGb`, `minCount`, `maxCount`. `maxCount >= minCount`. |
| `workloadsConfig.triggerer` | `object` | — | Triggerer: `cpu`, `memoryGb`, `count`. Critical for deferrable operators. |
| `workloadsConfig.dagProcessor` | `object` | — | Composer 3 only. DAG processor: `cpu`, `memoryGb`, `storageGb`, `count`. |
| `maintenanceWindow.startTime` | `string` | — | RFC3339 start (e.g., `2026-01-01T00:00:00Z`). Required if block is set. |
| `maintenanceWindow.endTime` | `string` | — | RFC3339 end. Required if block is set. |
| `maintenanceWindow.recurrence` | `string` | — | RFC5545 RRULE (e.g., `FREQ=WEEKLY;BYDAY=SA,SU`). Required if block is set. |
| `recoveryConfig.enabled` | `bool` | `false` | Enable scheduled environment snapshots. |
| `recoveryConfig.snapshotLocation` | `string` | — | GCS URI for snapshots. |
| `recoveryConfig.snapshotCreationSchedule` | `string` | — | Cron schedule (e.g., `0 4 * * *`). |
| `recoveryConfig.timeZone` | `string` | — | Timezone for cron (e.g., `UTC`, `America/Los_Angeles`). |
| `webServerNetworkAccessControl.allowedIpRanges` | `object[]` | — | IP allowlist for the Airflow UI. Each entry: `value` (CIDR, required), `description` (optional). |

## Examples

### With Software Configuration

An environment with a pinned Composer version and custom Python packages:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: my-data-pipelines
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: data-platform
    pulumi.planton.dev/stack.name: dev.GcpCloudComposerEnvironment.my-data-pipelines
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_SMALL
  softwareConfig:
    imageVersion: composer-2.9.7-airflow-2.9.3
    pypiPackages:
      apache-airflow-providers-google: ""
      numpy: ">=1.21"
    airflowConfigOverrides:
      webserver-dag_default_view: grid
```

### Production with Private Networking

A medium-sized environment with VPC peering, private endpoint, scaled workloads, and a maintenance window:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: prod-airflow
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: production
    pulumi.planton.dev/stack.name: prod.GcpCloudComposerEnvironment.prod-airflow
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_MEDIUM
  resilienceMode: HIGH_RESILIENCE
  nodeConfig:
    network:
      value: projects/my-gcp-project/global/networks/my-vpc
    subnetwork:
      value: projects/my-gcp-project/regions/us-central1/subnetworks/my-subnet
    serviceAccount:
      value: composer-sa@my-gcp-project.iam.gserviceaccount.com
  softwareConfig:
    imageVersion: composer-2.9.7-airflow-2.9.3
  privateEnvironmentConfig:
    enablePrivateEndpoint: true
    connectionType: VPC_PEERING
    masterIpv4CidrBlock: "172.16.0.0/28"
    cloudComposerNetworkIpv4CidrBlock: "10.0.48.0/20"
  workloadsConfig:
    scheduler:
      cpu: 2
      memoryGb: 7.5
      storageGb: 5
      count: 2
    webServer:
      cpu: 2
      memoryGb: 7.5
      storageGb: 5
    worker:
      cpu: 2
      memoryGb: 7.5
      storageGb: 5
      minCount: 2
      maxCount: 6
    triggerer:
      cpu: 1
      memoryGb: 1
      count: 2
  maintenanceWindow:
    startTime: "2026-01-01T00:00:00Z"
    endTime: "2026-01-01T12:00:00Z"
    recurrence: "FREQ=WEEKLY;BYDAY=SA,SU"
```

### Enterprise with CMEK and Recovery

A large environment with encryption, access control, and disaster recovery:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudComposerEnvironment
metadata:
  name: enterprise-airflow
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: enterprise
    pulumi.planton.dev/stack.name: prod.GcpCloudComposerEnvironment.enterprise-airflow
spec:
  projectId:
    value: my-gcp-project
  region: us-central1
  environmentSize: ENVIRONMENT_SIZE_LARGE
  resilienceMode: HIGH_RESILIENCE
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/my-ring/cryptoKeys/composer-key
  nodeConfig:
    network:
      value: projects/my-gcp-project/global/networks/my-vpc
    subnetwork:
      value: projects/my-gcp-project/regions/us-central1/subnetworks/my-subnet
    serviceAccount:
      value: composer-sa@my-gcp-project.iam.gserviceaccount.com
  softwareConfig:
    imageVersion: composer-2.9.7-airflow-2.9.3
  privateEnvironmentConfig:
    enablePrivateEndpoint: true
    connectionType: VPC_PEERING
    masterIpv4CidrBlock: "172.16.0.0/28"
    cloudSqlIpv4CidrBlock: "10.0.32.0/20"
    cloudComposerNetworkIpv4CidrBlock: "10.0.48.0/20"
  workloadsConfig:
    scheduler:
      cpu: 4
      memoryGb: 15
      storageGb: 10
      count: 2
    webServer:
      cpu: 4
      memoryGb: 15
      storageGb: 10
    worker:
      cpu: 4
      memoryGb: 15
      storageGb: 10
      minCount: 3
      maxCount: 10
    triggerer:
      cpu: 1
      memoryGb: 1
      count: 2
  maintenanceWindow:
    startTime: "2026-01-01T02:00:00Z"
    endTime: "2026-01-01T14:00:00Z"
    recurrence: "FREQ=WEEKLY;BYDAY=SA,SU"
  recoveryConfig:
    enabled: true
    snapshotLocation: gs://my-bucket/composer-snapshots
    snapshotCreationSchedule: "0 2 * * *"
    timeZone: UTC
  webServerNetworkAccessControl:
    allowedIpRanges:
      - value: "10.0.0.0/8"
        description: Internal VPN
      - value: "203.0.113.0/24"
        description: Office network
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `environment_id` | `string` | Fully qualified resource ID (`projects/{project}/locations/{region}/environments/{name}`) |
| `environment_name` | `string` | Short name of the environment |
| `airflow_uri` | `string` | URL of the Apache Airflow web UI |
| `dag_gcs_prefix` | `string` | Cloud Storage prefix for DAG uploads (e.g., `gs://{bucket}/dags`) |
| `gke_cluster` | `string` | Name of the underlying GKE cluster managed by Composer |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — Network for Composer 2.x VPC peering
- [GcpSubnetwork](/docs/catalog/gcp/subnetwork) — Subnetwork for node placement
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — Custom identity for Composer nodes
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — Customer-managed encryption key for CMEK
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) — Storage for DAGs, plugins, and snapshots
