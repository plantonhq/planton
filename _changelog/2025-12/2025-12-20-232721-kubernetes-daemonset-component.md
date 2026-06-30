# KubernetesDaemonSet Deployment Component

**Date**: December 20, 2025  
**Type**: Feature  
**Components**: Kubernetes Provider, API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added a new `KubernetesDaemonSet` deployment component that enables declarative deployment of DaemonSets to Kubernetes clusters. This component follows established patterns from `KubernetesDeployment` and `KubernetesCronJob`, providing DaemonSet-specific features like node selectors, tolerations, host path volumes, and privileged security contexts essential for node-level workloads.

## Problem Statement / Motivation

DaemonSets are a critical Kubernetes workload type for node-level services that must run on every (or a subset of) nodes in a cluster. Common use cases include:

- **Log Collection**: Fluentd, Fluent Bit, Filebeat running on every node
- **Node Monitoring**: Prometheus Node Exporter, Datadog agents
- **Network Plugins**: CNI implementations, network proxies
- **Storage Daemons**: Ceph, Longhorn distributed storage
- **Security Agents**: Falco, runtime security monitoring

### Pain Points

- No existing component for DaemonSet deployments in Planton
- Node-level workloads require specialized configuration (host paths, tolerations, privileged mode)
- Manual DaemonSet YAML is verbose and error-prone
- No unified way to deploy system daemons across infrastructure managed by Planton

## Solution / What's New

Introduced the `KubernetesDaemonSet` component with full proto API definitions and Pulumi IaC implementation.

### Component Architecture

```
KubernetesDaemonSet
├── Proto API (v1)
│   ├── spec.proto          # DaemonSet configuration
│   ├── api.proto           # KRM wiring (metadata/spec/status)
│   ├── stack_input.proto   # IaC module inputs
│   └── stack_outputs.proto # Deployment outputs
├── Pulumi Module
│   ├── main.go             # Resource orchestration
│   ├── daemonset.go        # DaemonSet creation
│   ├── locals.go           # Variable initialization
│   ├── secret.go           # Environment secrets
│   └── image_pull_secret.go
├── Tests
│   └── spec_test.go        # 17 validation tests
└── Documentation
    ├── README.md           # User guide
    ├── examples.md         # Usage examples
    └── docs/README.md      # Technical docs
```

### Key Features

| Feature | Description |
|---------|-------------|
| **Volume Mounts** | Host path volumes with configurable path types (Directory, File, Socket, etc.) |
| **Tolerations** | Schedule on tainted nodes (masters, GPU nodes, dedicated nodes) |
| **Node Selectors** | Target specific node types via label matching |
| **Security Context** | Privileged mode, Linux capabilities (add/drop), user/group IDs |
| **Update Strategy** | RollingUpdate with `maxUnavailable`/`maxSurge`, or OnDelete |
| **Health Probes** | Liveness, readiness, startup probes (HTTP, TCP, gRPC, exec) |
| **Command/Args** | Override container entrypoint and arguments |
| **Host Ports** | Map container ports to host network |

## Implementation Details

### Proto API Definition

The spec includes DaemonSet-specific configurations:

```protobuf
message KubernetesDaemonSetSpec {
  KubernetesClusterSelector target_cluster = 1;
  StringValueOrRef namespace = 2;
  bool create_namespace = 3;
  KubernetesDaemonSetContainer container = 4;
  map<string, string> node_selector = 5;
  repeated KubernetesDaemonSetToleration tolerations = 6;
  KubernetesDaemonSetUpdateStrategy update_strategy = 7;
  int32 min_ready_seconds = 8;
}
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/spec.proto`

### Validation Rules

Implemented CEL-based validations for:

- Image repo/tag required
- Port names: lowercase alphanumeric with hyphens
- Network protocol: must be TCP, UDP, or SCTP
- Toleration operators: Exists or Equal
- Toleration effects: NoSchedule, PreferNoSchedule, or NoExecute
- Update strategy types: RollingUpdate or OnDelete

### Pulumi Module

The module creates these Kubernetes resources:

1. **Namespace** (optional): When `create_namespace: true`
2. **Secret** (env-secrets): Contains environment secrets
3. **Secret** (image-pull): For private registry authentication
4. **DaemonSet**: The main workload controller

Key implementation in `daemonset.go`:

```go
// Build tolerations for node scheduling
tolerations := make(kubernetescorev1.TolerationArray, 0)
for _, t := range target.Spec.Tolerations {
    toleration := &kubernetescorev1.TolerationArgs{}
    if t.Key != "" {
        toleration.Key = pulumi.StringPtr(t.Key)
    }
    if t.Operator != "" {
        toleration.Operator = pulumi.StringPtr(t.Operator)
    }
    // ...
}
podSpecArgs.Tolerations = tolerations
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/iac/pulumi/module/daemonset.go`

### Resource Registration

Registered in `cloud_resource_kind.proto` as:

```protobuf
KubernetesDaemonSet = 841 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8sds"
}];
```

## Usage Examples

### Log Collector (Fluentd)

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
      volume_mounts:
        - name: varlog
          mount_path: /var/log
          host_path: /var/log
          read_only: true
  tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
      effect: NoSchedule
```

### Node Exporter (Prometheus)

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDaemonSet
metadata:
  name: node-exporter
spec:
  namespace:
    value: monitoring
  container:
    app:
      image:
        repo: prom/node-exporter
        tag: v1.7.0
      args:
        - --path.procfs=/host/proc
        - --path.sysfs=/host/sys
      volume_mounts:
        - name: proc
          mount_path: /host/proc
          host_path: /proc
          read_only: true
      ports:
        - name: metrics
          container_port: 9100
          network_protocol: TCP
  tolerations:
    - operator: Exists
```

## Benefits

### For Infrastructure Operators
- **Declarative Node Services**: Deploy log collectors, monitors, and agents consistently
- **GitOps Ready**: Version control DaemonSet configurations
- **Cross-Cluster Consistency**: Same manifest works across environments

### For Platform Engineers
- **Unified API**: Consistent KRM-style interface matching other Kubernetes components
- **Validation**: Proto-based validation catches errors before deployment
- **Extensibility**: Easy to add new DaemonSet-specific features

### For Developers
- **Simple Manifests**: Complex DaemonSet configs reduced to essential fields
- **Best Practices Built-In**: Proper labeling, resource naming, security context handling

## Impact

### Files Created

| Category | Files |
|----------|-------|
| Proto definitions | 4 files (spec, api, stack_input, stack_outputs) |
| Generated Go stubs | 4 files (*.pb.go) |
| Pulumi module | 6 Go files |
| Tests | 1 file (17 test cases) |
| Documentation | 4 markdown files |
| Build/Config | 3 files (Makefile, Pulumi.yaml, manifest.yaml) |

### Test Coverage

```
=== RUN   TestKubernetesDaemonSet
Running Suite: KubernetesDaemonSet Suite
Will run 17 of 17 specs
•••••••••••••••••
Ran 17 of 17 Specs in 0.013 seconds
SUCCESS! -- 17 Passed | 0 Failed
```

Test categories:
- Basic configuration validation
- Container image validation (repo/tag required)
- Port name and protocol validation
- Toleration operator/effect validation
- Update strategy type validation
- Node selector handling
- Volume mount configuration

## Related Work

- **KubernetesDeployment**: Similar pattern for Deployment workloads - referenced for consistency
- **KubernetesCronJob**: Similar pattern for CronJob workloads - referenced for simpler container spec
- **cloud_resource_kind.proto**: Extended with new `KubernetesDaemonSet` enum value (841)

## Future Enhancements

Potential additions for future iterations:

- **Pod Anti-Affinity**: Prevent scheduling on specific nodes
- **Priority Classes**: Pod priority for scheduling
- **Service Account**: Dedicated service account with RBAC
- **ConfigMaps**: Support for ConfigMap volume mounts
- **Init Containers**: Pre-start initialization containers
- **Terraform Module**: Parity with Pulumi implementation

---

**Status**: ✅ Production Ready  
**Timeline**: Single session implementation  
**Build**: All tests passing, full build successful

