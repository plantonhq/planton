# Standard Rook Ceph Operator

This preset deploys the Rook Ceph Operator with recommended default resources and default CSI settings. Rook enables Ceph distributed storage on Kubernetes, providing block, file, and object storage. This preset uses all default CSI driver settings (RBD and CephFS enabled, host networking on).

## When to Use

- You need distributed storage on Kubernetes via Ceph
- Default CSI driver configuration (RBD + CephFS enabled) meets your needs
- You will create a `KubernetesRookCephCluster` resource separately after the operator is running

## Key Configuration Choices

- **Namespace** (`rook-ceph`) -- the standard Rook Ceph namespace; Ceph clusters managed by this operator expect this namespace
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`200m` CPU, `128Mi` memory) -- matches the proto recommended defaults for the operator pod
- **Resource limits** (`500m` CPU, `512Mi` memory) -- conservative ceiling for the operator; not the storage daemons
- **CSI drivers** -- omitted, which means all defaults apply: RBD enabled, CephFS enabled, host networking enabled, 2 provisioner replicas

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

## Related Presets

- **02-production-with-csi** -- Explicitly configures CSI driver options for production clusters
