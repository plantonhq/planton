# KubernetesStrimziKafkaOperator

Deploys the Strimzi Kafka Operator to Kubernetes clusters, enabling declarative management of Apache Kafka clusters through Kubernetes Custom Resources.

## Overview

Strimzi is the production-standard Kubernetes operator for Apache Kafka. It automates deployment, scaling, configuration, and management of Kafka clusters, topics, and users.

### Why Strimzi?

- **CNCF Project**: Production-proven, vendor-neutral, 100% open source (Apache 2.0)
- **Battle-Tested**: Powers Kafka deployments at Red Hat, Bloomberg, and Fortune 500 companies
- **Comprehensive**: Manages clusters, topics, users, connectors, bridges via CRDs
- **GitOps-Native**: Full declarative configuration through Kubernetes resources

### What This Deploys

The operator itself (NOT Kafka clusters). After deployment, you create Kafka clusters using `Kafka` CRDs.

## Quick Start

### Basic Deployment

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: kafka-operator
spec:
  targetCluster:
    clusterName: "my-gke-cluster"
  namespace:
    value: "strimzi-kafka-operator"
  createNamespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

### Deploy

```bash
# Pulumi
cd iac/pulumi && pulumi up

# After operator deploys, create Kafka clusters with Kafka CRDs
```

## What Gets Created

- **Namespace**: `strimzi-kafka-operator` (if `createNamespace: true`)
- **Operator Deployment**: Watches for Kafka CRDs cluster-wide
- **CRDs**: Kafka, KafkaTopic, KafkaUser, KafkaConnect, KafkaBridge, KafkaMirrorMaker2
- **RBAC**: Cluster-wide permissions for operator

## Post-Installation: Creating Kafka Clusters

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: my-cluster
  namespace: kafka
spec:
  kafka:
    version: 3.6.0
    replicas: 3
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
    storage:
      type: jbod
      volumes:
        - id: 0
          type: persistent-claim
          size: 100Gi
          deleteClaim: false
  zookeeper:
    replicas: 3
    storage:
      type: persistent-claim
      size: 10Gi
      deleteClaim: false
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

## API Reference

### Spec

- **`targetCluster`** (optional): Kubernetes cluster to deploy operator
- **`namespace`** (required): Kubernetes namespace for the operator
- **`createNamespace`** (optional, default: `false`): Whether to create the namespace
  - `true`: Module creates the namespace with appropriate labels
  - `false`: Module assumes namespace already exists (must be pre-created)
- **`container.resources`**: Operator pod resource limits

### Defaults

```yaml
createNamespace: false  # Namespace must exist
container:
  resources:
    requests:
      cpu: "50m"
      memory: "100Mi"
    limits:
      cpu: "1000m"
      memory: "1Gi"
```

## Namespace Management

### When to use `createNamespace: true`
- **Simple deployments**: Let the module manage the full lifecycle
- **New clusters**: No pre-existing namespace configuration
- **Isolated operator**: Dedicated namespace for this operator only

### When to use `createNamespace: false`
- **Pre-configured namespaces**: Namespace has specific ResourceQuotas, NetworkPolicies, or LimitRanges
- **Shared namespaces**: Multiple operators in the same namespace
- **GitOps workflows**: Namespace managed by ArgoCD/Flux
- **Multi-tenant environments**: Namespace lifecycle controlled by platform team

**Note:** The protobuf boolean default is `false`, so you must explicitly set `createNamespace: true` if you want the module to create the namespace.

## Architecture

The operator runs in `strimzi-kafka-operator` namespace and watches ALL namespaces for Kafka CRDs (`watchAnyNamespace: true`). This enables multi-tenant Kafka deployments where teams create Kafka clusters in their own namespaces.

## Resource Requirements

- **Operator**: ~50Mi memory, <100m CPU baseline
- **Per Kafka Cluster**: Depends on cluster spec (typically 2Gi+ per broker)

## Best Practices

1. **One Operator Per Cluster**: Single operator manages all Kafka clusters
2. **Namespace Isolation**: Deploy Kafka clusters in separate namespaces
3. **Resource Planning**: Operator is lightweight; Kafka brokers are resource-intensive
4. **Monitoring**: Deploy Prometheus + Kafka Exporter for metrics
5. **TLS**: Enable TLS for production Kafka listeners

## Additional Resources

- **Research Documentation**: [docs/README.md](docs/README.md) - 32KB deep-dive into operator landscape, best practices, and production patterns
- **Examples**: [examples.md](examples.md) - Practical deployment scenarios
- **Strimzi Documentation**: https://strimzi.io/documentation/
- **Pulumi Module**: [iac/pulumi/README.md](iac/pulumi/README.md)

## License

Strimzi: Apache License 2.0  
This module: Part of Planton

