---
title: "Kafka"
description: "Kafka deployment documentation"
icon: "package"
order: 100
componentName: "kuberneteskafka"
---

# Kubernetes Kafka

Deploys an Apache Kafka cluster on Kubernetes using the Strimzi operator, with Zookeeper-based coordination, SCRAM-SHA-512 authentication, optional Schema Registry (Confluent), optional Kafka UI (Kowl), per-broker persistent storage, and external access via TLS-terminated load balancers with automatic DNS management.

## What Gets Created

When you deploy a KubernetesKafka resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Strimzi Kafka Cluster** — a `Kafka` custom resource with configurable broker replicas, resource limits, JBOD persistent storage, simple authorization, and SCRAM-SHA-512 authentication on all listeners
- **Zookeeper Ensemble** — co-deployed with the Kafka cluster, using persistent-claim storage for durable coordination state
- **Entity Operator** — Strimzi Topic Operator and User Operator for declarative topic and user management
- **Admin User** — a `KafkaUser` custom resource with SCRAM-SHA-512 credentials and super-user privileges, with the password stored in a Kubernetes Secret
- **Kafka Topics** — one `KafkaTopic` custom resource per entry in `kafkaTopics`, each with configurable partitions, replicas, and topic-level settings
- **Schema Registry** — a Confluent Schema Registry `Deployment` and `ClusterIP` Service, created only when `schemaRegistryContainer.isEnabled` is `true`; includes Gateway API ingress resources when ingress is enabled
- **Kafka UI (Kowl)** — a Kowl `Deployment`, `ConfigMap`, and `ClusterIP` Service, created only when `isDeployKafkaUi` is `true`; includes Gateway API ingress resources when ingress is enabled
- **TLS Certificates** — cert-manager `Certificate` resources for bootstrap server, schema registry, and Kowl hostnames, issued by a ClusterIssuer matching the ingress domain
- **Load Balancer Listeners** — public and private Strimzi load balancer listeners with per-broker advertised hostnames and external-dns annotations, created only when ingress is enabled
- **Gateway API Resources** — `Gateway` and `HTTPRoute` resources for Schema Registry and Kowl external access via Istio ingress

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Strimzi Kafka Operator** installed in the cluster (provides `Kafka`, `KafkaTopic`, and `KafkaUser` CRDs)
- **cert-manager** installed with a `ClusterIssuer` matching the ingress domain, required when ingress is enabled
- **external-dns** running in the cluster if enabling ingress with hostnames
- **Istio ingress gateway** deployed in the `istio-ingress` namespace, required for Schema Registry and Kowl external access
- **A StorageClass** available in the cluster for broker and Zookeeper persistent volumes

## Quick Start

Create a file `kafka.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKafka
metadata:
  name: my-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesKafka.my-kafka
spec:
  namespace: kafka
  createNamespace: true
```

Deploy:

```shell
planton apply -f kafka.yaml
```

This creates a single-broker Kafka cluster with a single Zookeeper node, 1Gi persistent storage for each, SCRAM-SHA-512 authentication, an admin super-user, and the Kowl UI enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Kafka deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `kafkaTopics` | `list` | `[]` | List of Kafka topics to create. Each entry accepts `name`, `partitions`, `replicas`, and `config`. |
| `kafkaTopics[].name` | `string` | — | Topic name. Must be 1-249 characters, start and end with an alphanumeric character, and contain only alphanumerics, `.`, `_`, or `-`. |
| `kafkaTopics[].partitions` | `int32` | `1` | Number of partitions for the topic. |
| `kafkaTopics[].replicas` | `int32` | `1` | Number of replicas for the topic. |
| `kafkaTopics[].config` | `map<string,string>` | See below | Topic-level configuration overrides. Defaults include `cleanup.policy: delete`, `retention.ms: 604800000`, `max.message.bytes: 2097164`, and others. |
| `brokerContainer.replicas` | `int32` | `1` | Number of Kafka broker pods. |
| `brokerContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each broker pod. |
| `brokerContainer.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each broker pod. |
| `brokerContainer.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each broker pod. |
| `brokerContainer.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each broker pod. |
| `brokerContainer.diskSize` | `string` | `1Gi` | Persistent volume size for each broker. Must be a valid Kubernetes quantity (e.g., `1Gi`, `30Gi`). |
| `zookeeperContainer.replicas` | `int32` | `1` | Number of Zookeeper pods. Use 3 or more for high availability (Raft consensus). |
| `zookeeperContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Zookeeper pod. |
| `zookeeperContainer.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Zookeeper pod. |
| `zookeeperContainer.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Zookeeper pod. |
| `zookeeperContainer.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Zookeeper pod. |
| `zookeeperContainer.diskSize` | `string` | `1Gi` | Persistent volume size for each Zookeeper node. Must be a valid Kubernetes quantity. |
| `schemaRegistryContainer.isEnabled` | `bool` | `false` | Deploys a Confluent Schema Registry alongside Kafka. |
| `schemaRegistryContainer.replicas` | `int32` | `1` | Number of Schema Registry pods. Only applies when `isEnabled` is `true`. |
| `schemaRegistryContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Schema Registry pod. |
| `schemaRegistryContainer.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Schema Registry pod. |
| `schemaRegistryContainer.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Schema Registry pod. |
| `schemaRegistryContainer.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Schema Registry pod. |
| `ingress.enabled` | `bool` | `false` | Enables external access via TLS-terminated load balancers for Kafka brokers, and Gateway API ingress for Schema Registry and Kowl. |
| `ingress.hostname` | `string` | — | Base hostname for external access (e.g., `kafka.example.com`). Required when `ingress.enabled` is `true`. Used to derive broker, Schema Registry, and Kowl hostnames. |
| `isDeployKafkaUi` | `bool` | `true` | Deploys the Kowl web UI for browsing topics, consumer groups, and messages. |

## Examples

### Development Kafka Cluster

A minimal single-broker Kafka cluster for development with reduced resources and no external access:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKafka
metadata:
  name: dev-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesKafka.dev-kafka
spec:
  namespace: dev-kafka
  createNamespace: true
  kafkaTopics:
    - name: events
    - name: notifications
  brokerContainer:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "100m"
        memory: "256Mi"
    diskSize: "5Gi"
  zookeeperContainer:
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "512Mi"
      requests:
        cpu: "50m"
        memory: "128Mi"
    diskSize: "1Gi"
```

### Production Kafka with Schema Registry and Ingress

A multi-broker production cluster with Schema Registry, Kafka UI, TLS ingress, and increased storage:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKafka
metadata:
  name: prod-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesKafka.prod-kafka
spec:
  namespace: kafka-prod
  kafkaTopics:
    - name: orders
      partitions: 12
      replicas: 3
    - name: user-events
      partitions: 6
      replicas: 3
      config:
        cleanup.policy: compact
        retention.ms: "-1"
    - name: dead-letter
      partitions: 3
      replicas: 3
      config:
        retention.ms: "2592000000"
  brokerContainer:
    replicas: 3
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "4Gi"
    diskSize: "100Gi"
  zookeeperContainer:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    diskSize: "10Gi"
  schemaRegistryContainer:
    isEnabled: true
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
  ingress:
    enabled: true
    hostname: kafka.prod.example.com
  isDeployKafkaUi: true
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKafka
metadata:
  name: shared-kafka
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesKafka.shared-kafka
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: messaging-ns
      field: spec.name
  kafkaTopics:
    - name: audit-log
      partitions: 6
      replicas: 2
      config:
        cleanup.policy: compact
        retention.ms: "-1"
  brokerContainer:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    diskSize: "50Gi"
  zookeeperContainer:
    replicas: 3
    diskSize: "5Gi"
  schemaRegistryContainer:
    isEnabled: true
  isDeployKafkaUi: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Kafka is deployed |
| `username` | `string` | SASL admin username for Kafka authentication |
| `passwordSecret.name` | `string` | Name of the Kubernetes Secret containing the admin password |
| `passwordSecret.key` | `string` | Key within the password Secret (always `password`) |
| `bootstrapServerExternalHostname` | `string` | Public hostname of the Kafka bootstrap server, only set when ingress is enabled |
| `bootstrapServerInternalHostname` | `string` | Internal (VPC-level) hostname of the Kafka bootstrap server, only set when ingress is enabled |
| `schemaRegistryExternalUrl` | `string` | Public HTTPS URL for the Schema Registry, empty when Schema Registry is not enabled |
| `schemaRegistryInternalUrl` | `string` | Internal HTTPS URL for the Schema Registry, empty when Schema Registry is not enabled |
| `kafkaUiExternalUrl` | `string` | Public HTTPS URL for the Kowl Kafka UI, only set when `isDeployKafkaUi` is `true` and ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — application deployments that produce to or consume from Kafka topics
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — often deployed alongside Kafka for event-sourced architectures with a relational read store
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — complementary caching layer for applications consuming Kafka events
