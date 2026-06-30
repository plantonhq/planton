# KubernetesRookCephCluster Pulumi Module

This Pulumi module deploys a Rook Ceph storage cluster on Kubernetes.

## Prerequisites

1. **Rook Operator Installed**: Deploy `KubernetesRookCephOperator` first
2. **Kubernetes Cluster**: With kubectl access configured
3. **Raw Block Devices**: On nodes for Ceph OSDs

## Usage

### Standalone Pulumi

```bash
# Build the module
make build

# Install required plugins
make install-pulumi-plugins

# Set the manifest path
export PLANTON_CLOUD_RESOURCE_MANIFEST_PATH=/path/to/manifest.yaml

# Deploy
pulumi up
```

### With Planton CLI

```bash
planton pulumi up --manifest manifest.yaml --stack org/project/env
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PLANTON_CLOUD_RESOURCE_MANIFEST_PATH` | Path to the KubernetesRookCephCluster manifest |
| `KUBERNETES_KUBECONFIG` | Path to kubeconfig (optional if using default) |

## Outputs

After deployment, the following outputs are available:

- `namespace` - Kubernetes namespace
- `helm_release_name` - Helm release name
- `ceph_cluster_name` - CephCluster resource name
- `block_pool_names` - List of CephBlockPool names
- `block_storage_class_names` - List of block StorageClass names
- `filesystem_names` - List of CephFilesystem names
- `filesystem_storage_class_names` - List of CephFS StorageClass names
- `object_store_names` - List of CephObjectStore names
- `object_storage_class_names` - List of object StorageClass names
- `dashboard_port_forward_command` - Command to access dashboard
- `dashboard_url` - Dashboard URL
- `dashboard_password_command` - Command to get dashboard password
- `toolbox_exec_command` - Command to access toolbox

## Debugging

Use the debug script for local testing:

```bash
export PLANTON_CLOUD_RESOURCE_MANIFEST_PATH=../hack/manifest.yaml
export KUBERNETES_KUBECONFIG=~/.kube/config
pulumi preview --stack dev
```

## Resources Created

- Kubernetes Namespace (if `create_namespace: true`)
- Helm Release for rook-ceph-cluster chart
  - CephCluster custom resource
  - CephBlockPool resources with StorageClasses
  - CephFilesystem resources with StorageClasses
  - CephObjectStore resources with StorageClasses
