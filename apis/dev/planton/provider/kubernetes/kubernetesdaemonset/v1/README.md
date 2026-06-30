# Overview

The **Kubernetes DaemonSet** API resource provides a standardized and streamlined way to deploy DaemonSets onto Kubernetes clusters. A DaemonSet ensures that all (or some) nodes run a copy of a pod. As nodes are added to the cluster, pods are added to them. As nodes are removed from the cluster, those pods are garbage collected.

## Purpose

DaemonSets are ideal for deploying system-level services that need to run on every node (or a subset of nodes) in the cluster. Common use cases include:

- **Log Collection**: Deploy log collectors like Fluentd, Fluent Bit, or Filebeat on every node
- **Node Monitoring**: Deploy monitoring agents like Prometheus Node Exporter or Datadog agents
- **Cluster Storage**: Deploy distributed storage daemons like Ceph or GlusterFS
- **Network Plugins**: Deploy CNI plugins or network proxies
- **Security Agents**: Deploy security scanning or compliance tools

The Kubernetes DaemonSet API resource aims to:

- **Standardize Deployments**: Offer a consistent interface for deploying DaemonSets
- **Simplify Configuration**: Consolidate all DaemonSet-related settings into one place
- **Node-Level Access**: Provide built-in support for host paths, tolerations, and privileged containers

## Key Features

### Namespace Management

- **Flexible Namespace Control**: Choose between creating a new dedicated namespace or deploying into an existing namespace via the `create_namespace` boolean flag
- **Isolated Deployments**: When `create_namespace` is `true`, each DaemonSet gets its own namespace with proper labeling
- **Multi-tenant Support**: When `create_namespace` is `false`, multiple DaemonSets can share the same namespace

### Container Specification

- **App Container Configuration**: Define the main application container, including:
  - **Container Image**: Set the container image and tag
  - **Resources**: Allocate CPU and memory resources
  - **Environment Variables and Secrets**: Manage configuration data and sensitive information
  - **Ports**: Configure container ports with optional host port mapping
  - **Volume Mounts**: Mount host paths for accessing node-level files/directories
  - **Security Context**: Configure privileged mode, user/group IDs, and Linux capabilities
  - **Command and Args**: Override container entrypoint and arguments

- **Health Probes**: Configure liveness, readiness, and startup probes

### Node Selection

- **Node Selector**: Use key-value pairs to constrain DaemonSet pods to specific nodes
- **Tolerations**: Allow pods to be scheduled on nodes with specific taints (e.g., master nodes)

### Update Strategy

- **RollingUpdate**: Progressively update pods with configurable `maxUnavailable` and `maxSurge`
- **OnDelete**: Only update pods when they are manually deleted

## Benefits

- **Consistency Across Nodes**: DaemonSets ensure your system services run on every applicable node
- **Automatic Scaling**: New nodes automatically get the DaemonSet pods
- **Node-Level Operations**: Built-in support for privileged access and host path mounts
- **Security**: Securely manage sensitive information with Kubernetes secrets
- **Flexibility**: Support for node selectors and tolerations for fine-grained control

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDaemonSet
metadata:
  name: fluentd-logger
spec:
  namespace:
    value: logging
  create_namespace: true
  container:
    app:
      image:
        repo: fluent/fluentd-kubernetes-daemonset
        tag: v1.16-debian-elasticsearch8
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 200Mi
      volume_mounts:
        - name: varlog
          mount_path: /var/log
          host_path: /var/log
          read_only: true
        - name: containers
          mount_path: /var/lib/docker/containers
          host_path: /var/lib/docker/containers
          read_only: true
      ports:
        - name: metrics
          container_port: 24231
          network_protocol: TCP
  tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
      effect: NoSchedule
  update_strategy:
    type: RollingUpdate
    rolling_update:
      max_unavailable: "1"
```

