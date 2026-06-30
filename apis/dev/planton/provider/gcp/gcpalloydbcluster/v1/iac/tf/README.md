# GcpAlloydbCluster — Terraform Module

Terraform implementation for the GcpAlloydbCluster Planton component. Provisions an AlloyDB cluster with a bundled primary instance.

## Resources Created

- `google_alloydb_cluster` — AlloyDB cluster (backup, encryption, network config)
- `google_alloydb_instance` — Primary instance (compute, availability, query insights)

## Usage

```hcl
module "alloydb_cluster" {
  source = "."

  metadata = {
    name = "my-alloydb"
  }

  spec = {
    project_id   = "my-gcp-project"
    cluster_name = "my-alloydb-cluster"
    location     = "us-central1"
    network      = "projects/my-gcp-project/global/networks/default"

    primary_instance = {
      instance_id       = "my-primary"
      cpu_count         = 4
      availability_type = "REGIONAL"
    }
  }
}
```

## Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `metadata` | Yes | Resource metadata (name, org, env, id) |
| `spec` | Yes | GcpAlloydbCluster specification |
| `provider_config` | No | GCP provider configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `cluster_id` | Fully qualified cluster resource name |
| `cluster_name` | Short name of the cluster |
| `primary_instance_ip` | Private IP address of the primary instance |
| `primary_instance_name` | Fully qualified primary instance resource name |
| `database_version` | Computed database engine version |
| `state` | Current state of the cluster |

## Notes

- The VPC must have Private Service Access configured before creating the cluster.
- `cluster_name`, `location`, `network`, and `kms_key_name` are immutable after creation.
- `primary_instance.instance_id` is immutable after creation.
- Provider version `~> 6.0` (google) is required.
