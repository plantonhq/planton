---
title: "Kafka"
description: "Kafka deployment documentation"
icon: "package"
order: 100
componentName: "confluentkafka"
---

# Confluent Kafka

Deploys a Confluent Cloud Kafka cluster with configurable cluster type, multi-zone availability, and optional private networking. Supports Basic, Standard, Enterprise, and Dedicated cluster tiers across AWS, Azure, and GCP regions.

## What Gets Created

When you deploy a ConfluentKafka resource, Planton provisions:

- **Confluent Cloud Kafka Cluster** — a `confluent_kafka_cluster` resource of the specified type (Basic, Standard, Enterprise, or Dedicated), placed in the given cloud provider region and associated with a Confluent Cloud environment
- **Network Association** — created only when `networkConfig` is provided, associates the cluster with a pre-existing Confluent Cloud network for private connectivity (PrivateLink on AWS, Private Link on Azure, Private Service Connect on GCP)

## Prerequisites

- **Confluent Cloud credentials** configured via environment variables (`CONFLUENT_CLOUD_API_KEY`, `CONFLUENT_CLOUD_API_SECRET`) or Planton provider config
- **A Confluent Cloud environment** — the `environmentId` of an existing environment where the cluster will be created
- **A Confluent Cloud network** if enabling private networking via `networkConfig` — must be pre-created in the same environment
- **Sufficient CKU quota** if deploying a Dedicated cluster

## Quick Start

Create a file `confluent-kafka.yaml`:

```yaml
apiVersion: confluent.planton.dev/v1
kind: ConfluentKafka
metadata:
  name: my-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ConfluentKafka.my-kafka
spec:
  cloud: AWS
  region: us-east-2
  availability: SINGLE_ZONE
  environmentId: env-abc123
```

Deploy:

```shell
planton apply -f confluent-kafka.yaml
```

This creates a Standard Kafka cluster in a single availability zone on AWS `us-east-2`, associated with the specified Confluent Cloud environment.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `cloud` | `string` | Cloud provider where the cluster is deployed. | Must be one of: `AWS`, `AZURE`, `GCP` |
| `region` | `string` | Cloud-specific region identifier (e.g., `us-east-2`, `us-central1`, `eastus`). | Minimum length: 1 |
| `availability` | `string` | High availability configuration. `SINGLE_ZONE`: single AZ, no SLA. `MULTI_ZONE`: multi-AZ, 99.99% SLA. `LOW` and `HIGH` are legacy values for Basic clusters. | Must be one of: `SINGLE_ZONE`, `MULTI_ZONE`, `LOW`, `HIGH` |
| `environmentId` | `string` | ID of the Confluent Cloud environment (parent container for the cluster). | Minimum length: 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterType` | `string` | `STANDARD` | Cluster deployment type. `BASIC`: multi-tenant, single-zone only, public internet. `STANDARD`: multi-tenant, elastic scaling, public internet. `ENTERPRISE`: multi-tenant, elastic scaling, supports private networking. `DEDICATED`: single-tenant, provisioned CKU capacity, supports private networking. |
| `dedicatedConfig.cku` | `int` | — | Confluent Kafka Units for Dedicated clusters. Required when `clusterType` is `DEDICATED`. Minimum: 1. |
| `networkConfig.networkId` | `string` | — | ID of a pre-existing Confluent Cloud network resource for private connectivity. Only available for Enterprise and Dedicated cluster types. |
| `displayName` | `string` | `metadata.name` | Human-readable name for the cluster. If not specified, defaults to `metadata.name`. |

## Examples

### Basic Development Cluster

A low-cost single-zone cluster for development and testing:

```yaml
apiVersion: confluent.planton.dev/v1
kind: ConfluentKafka
metadata:
  name: dev-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ConfluentKafka.dev-kafka
spec:
  cloud: GCP
  region: us-central1
  availability: SINGLE_ZONE
  environmentId: env-dev456
  clusterType: BASIC
  displayName: "Dev Kafka Cluster"
```

### Standard Multi-Zone Production Cluster

A production-grade cluster with multi-zone availability for high availability:

```yaml
apiVersion: confluent.planton.dev/v1
kind: ConfluentKafka
metadata:
  name: prod-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ConfluentKafka.prod-kafka
spec:
  cloud: AWS
  region: us-east-1
  availability: MULTI_ZONE
  environmentId: env-prod789
  clusterType: STANDARD
  displayName: "Production Event Bus"
```

### Dedicated Cluster with Private Networking

A single-tenant Dedicated cluster with provisioned capacity and private network connectivity, suitable for regulated workloads requiring network isolation:

```yaml
apiVersion: confluent.planton.dev/v1
kind: ConfluentKafka
metadata:
  name: secure-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ConfluentKafka.secure-kafka
spec:
  cloud: AZURE
  region: eastus
  availability: MULTI_ZONE
  environmentId: env-secure012
  clusterType: DEDICATED
  dedicatedConfig:
    cku: 2
  networkConfig:
    networkId: n-abc123
  displayName: "Secure Event Platform"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Provider-assigned unique ID for the Kafka cluster |
| `bootstrap_endpoint` | `string` | Bootstrap endpoint for Kafka client connections (e.g., `SASL_SSL://pkc-00000.us-central1.gcp.confluent.cloud:9092`) |
| `crn` | `string` | Confluent Resource Name for RBAC and API references (e.g., `crn://confluent.cloud/organization=.../environment=.../cloud-cluster=...`) |
| `rest_endpoint` | `string` | REST endpoint for the Kafka cluster (e.g., `https://pkc-00000.us-central1.gcp.confluent.cloud:443`) |

## Related Components

No other Planton components have direct foreign key references to ConfluentKafka. This component is typically deployed alongside application workloads that produce or consume Kafka messages.
