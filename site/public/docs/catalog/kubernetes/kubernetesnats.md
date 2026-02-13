---
title: "NATS"
description: "NATS deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesnats"
---

# Kubernetes NATS

Deploys a NATS messaging cluster on Kubernetes using the official NATS Helm chart, with optional JetStream persistence, authentication (bearer-token or basic-auth), TLS encryption, external ingress via LoadBalancer, and declarative stream/consumer management through the NACK JetStream controller.

## What Gets Created

When you deploy a KubernetesNats resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release (NATS)** — deploys a NATS server cluster via the official NATS Helm chart with configurable replicas, JetStream file storage, resource limits, and disk size
- **Auth Secret** — a Kubernetes Secret storing a randomly generated bearer token or basic-auth credentials, created when authentication is enabled
- **No-Auth User Secret** — an additional Kubernetes Secret for the unauthenticated user, created when basic-auth is enabled with `noAuthUser.enabled` set to `true`
- **TLS Secret** — a self-signed certificate (RSA 2048-bit, 5-year validity) stored as a `kubernetes.io/tls` Secret, created when `tlsEnabled` is `true`
- **LoadBalancer Service** — created when `ingress.enabled` is `true`, exposes NATS on port 4222 with an `external-dns.alpha.kubernetes.io/hostname` annotation for automatic DNS record creation
- **NACK CRDs** — JetStream custom resource definitions fetched from the official NACK GitHub release, deployed when the NACK controller is enabled
- **NACK Controller Helm Release** — deploys the NATS Controllers for Kubernetes operator, which reconciles JetStream Stream and Consumer CRDs to the NATS server
- **JetStream Stream CRs** — Kubernetes custom resources representing JetStream streams, each with configurable subjects, storage, retention, and limits
- **JetStream Consumer CRs** — Kubernetes custom resources representing JetStream consumers attached to streams, with configurable delivery, acknowledgment, and replay policies

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A StorageClass** available in the cluster if JetStream persistence is enabled (most managed Kubernetes clusters provide a default)
- **external-dns** running in the cluster if enabling ingress with a hostname

## Quick Start

Create a file `nats.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNats
metadata:
  name: my-nats
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesNats.my-nats
spec:
  namespace: messaging
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f nats.yaml
```

This creates a single-replica NATS server with JetStream enabled, a 10Gi PersistentVolumeClaim, default resource limits (1000m CPU, 2Gi memory), and the nats-box utility pod for debugging.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the NATS deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `serverContainer.replicas` | `int32` | `1` | Number of NATS server replicas. Use an odd value for quorum in clustered mode. Must be greater than 0. |
| `serverContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each NATS pod. |
| `serverContainer.resources.limits.memory` | `string` | `2Gi` | Maximum memory allocation for each NATS pod. |
| `serverContainer.resources.requests.cpu` | `string` | `100m` | Minimum guaranteed CPU for each NATS pod. |
| `serverContainer.resources.requests.memory` | `string` | `256Mi` | Minimum guaranteed memory for each NATS pod. |
| `serverContainer.diskSize` | `string` | `10Gi` | PVC size for JetStream file storage. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). Recommended default: `1Gi`. |
| `disableJetStream` | `bool` | `false` | When `true`, disables JetStream persistence entirely. No file store PVC is created. |
| `auth.enabled` | `bool` | `false` | Enables authentication for the NATS cluster. |
| `auth.scheme` | `enum` | — | Authentication scheme. Valid values: `bearer_token`, `basic_auth`. |
| `auth.noAuthUser.enabled` | `bool` | `false` | Enables an unauthenticated user alongside authenticated users. Only applies when `auth.scheme` is `basic_auth`. |
| `auth.noAuthUser.publishSubjects` | `string[]` | — | Subjects the unauthenticated user may publish to. At least one subject must be specified when `noAuthUser.enabled` is `true`. |
| `tlsEnabled` | `bool` | `false` | Generates a self-signed TLS certificate and configures NATS to use it. |
| `ingress.enabled` | `bool` | `false` | Creates a LoadBalancer Service exposing NATS on port 4222 with external-dns annotations. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `nats.example.com`). Configured automatically via external-dns. Required when `ingress.enabled` is `true`. |
| `disableNatsBox` | `bool` | `false` | When `true`, skips deployment of the nats-box utility pod. |
| `nackController.enabled` | `bool` | `false` | Deploys the NACK JetStream controller for managing streams and consumers via Kubernetes CRDs. |
| `nackController.enableControlLoop` | `bool` | `false` | Enables control-loop mode for the NACK controller. Required for KeyValue and ObjectStore support. |
| `nackController.helmChartVersion` | `string` | `0.31.1` | NACK Helm chart version. |
| `nackController.appVersion` | `string` | `0.21.1` | NACK app version (GitHub release tag). Used for fetching CRDs. Differs from chart version. |
| `streams[].name` | `string` | — | Unique stream name. 1-255 characters, alphanumeric with `-`, `_`, `.` allowed. |
| `streams[].subjects` | `string[]` | — | Subjects the stream captures. Supports wildcards (e.g., `orders.*`, `events.>`). At least one required. |
| `streams[].storage` | `enum` | `memory` | Storage backend: `file` (persistent) or `memory` (ephemeral). |
| `streams[].replicas` | `int32` | — | Number of stream replicas (1-5). Odd values recommended for quorum. |
| `streams[].retention` | `enum` | `limits` | Retention policy: `limits`, `interest`, or `workqueue`. |
| `streams[].maxAge` | `string` | — | Maximum message age (e.g., `24h`, `7d`). Empty for unlimited. |
| `streams[].maxBytes` | `int64` | — | Maximum stream size in bytes. `-1` for unlimited. |
| `streams[].maxMsgs` | `int64` | — | Maximum number of messages. `-1` for unlimited. |
| `streams[].maxMsgSize` | `int32` | — | Maximum message size in bytes. `-1` for unlimited. |
| `streams[].maxConsumers` | `int32` | — | Maximum number of consumers. `-1` for unlimited. |
| `streams[].discard` | `enum` | `old` | Discard policy when limits are reached: `old` or `new_msgs`. |
| `streams[].description` | `string` | — | Description of the stream. |
| `streams[].consumers[].durableName` | `string` | — | Durable consumer name. Must be unique within the stream. 1-255 characters. |
| `streams[].consumers[].deliverPolicy` | `enum` | `all` | Delivery policy: `all`, `last`, or `new_msgs`. |
| `streams[].consumers[].ackPolicy` | `enum` | `none` | Acknowledgment policy: `none`, `all`, or `explicit`. |
| `streams[].consumers[].filterSubject` | `string` | — | Subject filter with wildcard support. Only matching messages are delivered. |
| `streams[].consumers[].deliverSubject` | `string` | — | Deliver subject for push-based consumers. Omit for pull-based. |
| `streams[].consumers[].deliverGroup` | `string` | — | Queue group name for load balancing across multiple consumer instances. |
| `streams[].consumers[].maxAckPending` | `int32` | — | Maximum number of unacknowledged messages. |
| `streams[].consumers[].maxDeliver` | `int32` | — | Maximum delivery attempts. `-1` for unlimited. |
| `streams[].consumers[].ackWait` | `string` | — | Time to wait for acknowledgment (e.g., `30s`, `1m`). |
| `streams[].consumers[].replayPolicy` | `enum` | `instant` | Replay policy: `original` (original rate) or `instant` (as fast as possible). |
| `streams[].consumers[].description` | `string` | — | Description of the consumer. |
| `natsHelmChartVersion` | `string` | `2.12.3` | NATS Helm chart version. Available versions: `helm search repo nats/nats --versions`. |

## Examples

### Development NATS with Defaults

A single-replica NATS server for development with JetStream enabled and default settings:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNats
metadata:
  name: dev-nats
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesNats.dev-nats
spec:
  namespace: dev
  createNamespace: true
  serverContainer:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "50m"
        memory: "128Mi"
    diskSize: "1Gi"
```

### Production Cluster with Authentication and TLS

A three-node NATS cluster with basic-auth, TLS encryption, and external access:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNats
metadata:
  name: prod-nats
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesNats.prod-nats
spec:
  namespace: messaging
  serverContainer:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    diskSize: "50Gi"
  auth:
    enabled: true
    scheme: basic_auth
  tlsEnabled: true
  ingress:
    enabled: true
    hostname: nats.example.com
```

### Event-Driven Architecture with Streams and Consumers

A NATS cluster with the NACK controller managing JetStream streams and consumers declaratively:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNats
metadata:
  name: event-bus
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesNats.event-bus
spec:
  namespace: events
  createNamespace: true
  serverContainer:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    diskSize: "100Gi"
  auth:
    enabled: true
    scheme: basic_auth
  nackController:
    enabled: true
    enableControlLoop: true
  streams:
    - name: ORDERS
      subjects:
        - "orders.>"
      storage: file
      replicas: 3
      retention: limits
      maxAge: "7d"
      maxBytes: -1
      discard: old
      consumers:
        - durableName: order-processor
          deliverPolicy: all
          ackPolicy: explicit
          filterSubject: "orders.created"
          maxAckPending: 1000
          maxDeliver: 5
          ackWait: "30s"
        - durableName: order-analytics
          deliverPolicy: all
          ackPolicy: none
          description: "Analytics consumer, no ack required"
    - name: NOTIFICATIONS
      subjects:
        - "notify.*"
      storage: memory
      replicas: 1
      retention: interest
      consumers:
        - durableName: email-sender
          deliverPolicy: new_msgs
          ackPolicy: explicit
          filterSubject: "notify.email"
          ackWait: "1m"
          maxDeliver: 3
```

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNats
metadata:
  name: my-nats
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesNats.my-nats
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: app-namespace
      field: spec.name
  serverContainer:
    replicas: 3
    diskSize: "20Gi"
  auth:
    enabled: true
    scheme: bearer_token
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where NATS is deployed |
| `clientUrlInternal` | `string` | Cluster-internal NATS URL (e.g., `nats://my-nats.messaging.svc.cluster.local:4222`) |
| `clientUrlExternal` | `string` | External NATS URL, only set when ingress is enabled |
| `authTokenSecret.name` | `string` | Name of the Kubernetes Secret storing the auth credentials |
| `authTokenSecret.key` | `string` | Key within the auth Secret (`token` for bearer-token, `password` for basic-auth) |
| `jetStreamDomain` | `string` | JetStream domain configured for the cluster, blank when JetStream is disabled |
| `metricsEndpoint` | `string` | Prometheus metrics endpoint (e.g., `http://nats-prom.messaging.svc.cluster.local:7777/metrics`) |
| `tlsSecret.name` | `string` | Name of the Kubernetes Secret containing the TLS certificate and key, blank when TLS is disabled |
| `tlsSecret.key` | `string` | Key within the TLS Secret (`tls.crt`), blank when TLS is disabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that connect to NATS as a messaging backend
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — often deployed alongside NATS for caching in event-driven architectures
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — persistent storage for applications consuming NATS events
