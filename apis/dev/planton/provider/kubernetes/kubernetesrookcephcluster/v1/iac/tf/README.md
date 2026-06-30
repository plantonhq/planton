# KubernetesRookCephCluster Terraform Module

This Terraform module deploys a Rook Ceph storage cluster on Kubernetes.

## Prerequisites

1. **Rook Operator Installed**: Deploy `KubernetesRookCephOperator` first
2. **Kubernetes Cluster**: With kubectl access configured
3. **Raw Block Devices**: On nodes for Ceph OSDs

## Usage

### Basic Usage

```hcl
module "ceph_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "my-ceph-cluster"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    namespace        = "rook-ceph"
    create_namespace = true
    
    block_pools = [{
      name            = "ceph-blockpool"
      replicated_size = 3
      storage_class = {
        name       = "ceph-block"
        is_default = true
      }
    }]
  }
}
```

### Full Configuration

```hcl
module "ceph_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "production-ceph"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    namespace          = "rook-ceph"
    create_namespace   = true
    operator_namespace = "rook-ceph"
    helm_chart_version = "v1.16.6"

    ceph_image = {
      repository        = "quay.io/ceph/ceph"
      tag               = "v19.2.3"
      allow_unsupported = false
    }

    cluster = {
      data_dir_host_path = "/var/lib/rook"
      mon = {
        count                   = 3
        allow_multiple_per_node = false
      }
      mgr = {
        count                   = 2
        allow_multiple_per_node = false
      }
      storage = {
        use_all_nodes   = true
        use_all_devices = true
      }
    }

    block_pools = [{
      name            = "ceph-blockpool"
      failure_domain  = "host"
      replicated_size = 3
      storage_class = {
        enabled               = true
        name                  = "ceph-block"
        is_default            = true
        reclaim_policy        = "Delete"
        allow_volume_expansion = true
        volume_binding_mode   = "Immediate"
      }
    }]

    filesystems = [{
      name                          = "ceph-filesystem"
      metadata_pool_replicated_size = 3
      data_pool_replicated_size     = 3
      active_mds_count              = 1
      active_standby                = true
      storage_class = {
        enabled = true
        name    = "ceph-filesystem"
      }
    }]

    object_stores = [{
      name                            = "ceph-objectstore"
      metadata_pool_replicated_size   = 3
      data_pool_erasure_data_chunks   = 2
      data_pool_erasure_coding_chunks = 1
      gateway_port                    = 80
      gateway_instances               = 1
      storage_class = {
        enabled = true
        name    = "ceph-bucket"
      }
    }]

    enable_toolbox   = true
    enable_dashboard = true
    enable_monitoring = false
  }
}
```

## Inputs

| Name | Description | Type | Default |
|------|-------------|------|---------|
| metadata | Resource metadata | object | required |
| spec | Cluster specification | object | required |

See `variables.tf` for full specification details.

## Outputs

| Name | Description |
|------|-------------|
| namespace | Kubernetes namespace |
| helm_release_name | Helm release name |
| ceph_cluster_name | CephCluster resource name |
| block_pool_names | List of CephBlockPool names |
| block_storage_class_names | List of block StorageClass names |
| filesystem_names | List of CephFilesystem names |
| filesystem_storage_class_names | List of CephFS StorageClass names |
| object_store_names | List of CephObjectStore names |
| object_storage_class_names | List of object StorageClass names |
| dashboard_port_forward_command | Command to access dashboard |
| dashboard_url | Dashboard URL |
| dashboard_password_command | Command to get dashboard password |
| toolbox_exec_command | Command to access toolbox |

## Resources Created

- Kubernetes Namespace (if `create_namespace: true`)
- Helm Release for rook-ceph-cluster chart
  - CephCluster custom resource
  - CephBlockPool resources with StorageClasses
  - CephFilesystem resources with StorageClasses
  - CephObjectStore resources with StorageClasses
