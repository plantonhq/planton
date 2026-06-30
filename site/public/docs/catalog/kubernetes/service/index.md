---
title: "Service"
description: "Service deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesservice"
---

# Kubernetes Service

Deploys a standalone Kubernetes Service for service discovery, load balancing, and external access to workloads running in a Kubernetes cluster. Supports all four service types (ClusterIP, NodePort, LoadBalancer, ExternalName), headless mode, session affinity, external traffic policies, and cloud-provider-specific annotations for fine-grained load balancer control.

## What Gets Created

When you deploy a KubernetesService resource, Planton provisions:

- **Service** — a Kubernetes Service in the specified namespace with the configured type, port mappings, pod selector, and optional headless mode (`clusterIP: None`)
- **Labels** — standard Planton governance labels (`managed-by`, `resource`, `resource-kind`) merged with any user-provided labels
- **Annotations** — user-provided annotations passed through to the Service resource, enabling cloud-provider-specific load balancer configuration (e.g., AWS NLB, GCP NEG, Azure internal LB)

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **An existing namespace** in the target cluster where the service will be created
- **Target pods** with labels matching the `selector` field (not required for ExternalName or selectorless services)
- **Cloud provider load balancer support** if using `loadBalancer` type — the cluster must have a cloud controller manager or load balancer controller installed

## Quick Start

Create a file `service.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesService
metadata:
  name: my-service
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesService.my-service
spec:
  namespace: my-namespace
  name: my-service
  selector:
    app: my-app
  ports:
    - name: http
      port: 80
      targetPort: "8080"
```

Deploy:

```shell
planton apply -f service.yaml
```

This creates a ClusterIP Service named `my-service` in the `my-namespace` namespace, routing TCP traffic on port 80 to port 8080 on pods matching label `app: my-app`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `spec.namespace` | `string` | Kubernetes namespace where the service will be created. Must already exist. | 1–63 characters, valid DNS label (lowercase alphanumeric and hyphens, no leading/trailing hyphens) |
| `spec.name` | `string` | Name of the Kubernetes Service resource. Used as `metadata.name` on the created Service. | 1–63 characters, valid DNS label |
| `spec.ports` | `KubernetesServicePort[]` | Ports exposed by the service. At least one port is required for all service types except ExternalName. | Non-empty for non-ExternalName types |
| `spec.ports[].port` | `int32` | The port number exposed by the service. This is the port clients use to access the service. | 1–65535 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `spec.targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `spec.targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `spec.labels` | `map<string, string>` | `{}` | Additional labels merged with standard Planton labels for governance and discoverability. |
| `spec.annotations` | `map<string, string>` | `{}` | Annotations applied to the Service resource. Used for cloud-provider-specific load balancer configuration. |
| `spec.type` | `enum` | `cluster_ip` | Service type. Valid values: `cluster_ip`, `node_port`, `load_balancer`, `external_name`. |
| `spec.selector` | `map<string, string>` | `{}` | Label selector for pods this service routes traffic to. Not required for ExternalName services or services with manually managed endpoints. |
| `spec.headless` | `bool` | `false` | When `true`, creates a headless service (`clusterIP: None`). DNS queries return pod IPs directly. Cannot be combined with `node_port` or `load_balancer` types. |
| `spec.externalDnsName` | `string` | — | External DNS name for ExternalName-type services (e.g., `my-database.example.com`). Required when `type` is `external_name`. |
| `spec.externalTrafficPolicy` | `enum` | `cluster` | Controls how external traffic is routed. `cluster` distributes across all nodes (may add a hop). `local` routes only to node-local endpoints (preserves source IP). Only applicable for `node_port` and `load_balancer` types. |
| `spec.sessionAffinity` | `enum` | `none` | Session affinity configuration. `none` routes each request to any backend. `client_ip` routes requests from the same client IP to the same pod. |
| `spec.loadBalancerSourceRanges` | `string[]` | `[]` | CIDR ranges restricting load balancer access (e.g., `203.0.113.0/24`). Only applicable for `load_balancer` type. Empty means accessible from any source. |
| `spec.ports[].name` | `string` | — | Optional port name. Must be unique within the port list. Required when the service exposes more than one port. Max 63 characters. |
| `spec.ports[].protocol` | `enum` | `TCP` | IP protocol for this port. Valid values: `TCP`, `UDP`, `SCTP`. |
| `spec.ports[].targetPort` | `string` | Same as `port` | Target port on the pod. Can be a numeric port (e.g., `"8080"`) or a named port (e.g., `"http"`) defined in the container's port list. |
| `spec.ports[].nodePort` | `int32` | Auto-allocated | Port on each node for NodePort and LoadBalancer services. Must be in range 30000–32767 when specified. Only applicable for `node_port` and `load_balancer` types. |

## Examples

### Basic ClusterIP Service

A simple internal service routing HTTP traffic to backend pods:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesService
metadata:
  name: backend-api
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesService.backend-api
spec:
  namespace: backend
  name: backend-api
  selector:
    app: backend-api
    tier: api
  ports:
    - name: http
      port: 80
      targetPort: "8080"
    - name: grpc
      port: 9090
      targetPort: "9090"
```

### LoadBalancer with Source IP Restrictions

An externally accessible service using a cloud load balancer with traffic restricted to specific CIDR ranges and client source IP preservation:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesService
metadata:
  name: public-gateway
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesService.public-gateway
spec:
  namespace: ingress
  name: public-gateway
  type: load_balancer
  selector:
    app: gateway
  ports:
    - name: https
      port: 443
      targetPort: "8443"
    - name: http
      port: 80
      targetPort: "8080"
  externalTrafficPolicy: local
  sessionAffinity: client_ip
  loadBalancerSourceRanges:
    - "10.0.0.0/8"
    - "203.0.113.0/24"
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
```

### Full-Featured Multi-Protocol Service with NodePort

A production service exposing multiple protocols on explicit node ports, with session affinity and custom labels:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesService
metadata:
  name: media-server
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesService.media-server
spec:
  targetCluster:
    clusterKind: AwsEksCluster
    clusterName: prod-cluster
  namespace: media
  name: media-server
  type: node_port
  selector:
    app: media-server
    version: v3
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: "8080"
      nodePort: 30080
    - name: rtsp
      protocol: TCP
      port: 554
      targetPort: "8554"
      nodePort: 30554
    - name: stream
      protocol: UDP
      port: 5004
      targetPort: "5004"
      nodePort: 30504
  externalTrafficPolicy: local
  sessionAffinity: client_ip
  labels:
    team: media-platform
    cost-center: streaming

```

### ExternalName Service

A CNAME alias pointing to an external database endpoint, with no pod selector or ports:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesService
metadata:
  name: external-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesService.external-db
spec:
  namespace: backend
  name: external-db
  type: external_name
  externalDnsName: prod-db.c9kl3xq2.us-east-1.rds.amazonaws.com
```

### Headless Service for StatefulSet

A headless service for direct pod addressing, commonly used with StatefulSets:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesService
metadata:
  name: cassandra
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesService.cassandra
spec:
  namespace: data
  name: cassandra
  headless: true
  selector:
    app: cassandra
  ports:
    - name: cql
      port: 9042
      targetPort: "9042"
    - name: inter-node
      port: 7000
      targetPort: "7000"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `serviceName` | `string` | Name of the created Kubernetes Service |
| `namespace` | `string` | Namespace in which the service was created |
| `type` | `string` | Type of the created service (`ClusterIP`, `NodePort`, `LoadBalancer`, `ExternalName`) |
| `clusterIp` | `string` | Cluster-internal IP address assigned to the service. Empty for headless and ExternalName services. |
| `loadBalancerIngress` | `string` | External IP or hostname assigned by the cloud load balancer. Only populated for LoadBalancer-type services after provisioning. |
| `internalDnsName` | `string` | Fully qualified internal DNS name (e.g., `my-service.my-namespace.svc.cluster.local`) |

## Related Components

- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — deploys the workload pods that this service routes traffic to
- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace for the service
- [KubernetesStatefulSet](/docs/catalog/kubernetes/statefulset) — commonly used with headless services for stable pod identity and direct addressing
