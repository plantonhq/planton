# Terraform Module to Deploy AwsDocumentDb

This Terraform module deploys an AWS DocumentDB cluster (MongoDB-compatible document database) using the Planton API.

## Requirements

- Terraform >= 1.0
- AWS Provider ~> 5.82
- Valid AWS credentials configured

## Resources Created

- `aws_docdb_cluster` - DocumentDB cluster
- `aws_docdb_cluster_instance` - Cluster instances (count based on `instance_count`)
- `aws_docdb_subnet_group` - DB subnet group (when `subnet_ids` provided)
- `aws_security_group` - Security group (when `security_group_ids` or `allowed_cidr_blocks` provided)
- `aws_docdb_cluster_parameter_group` - Parameter group (when `cluster_parameters` provided)

## Usage

### With Planton CLI

```shell
# Plan
planton tofu plan --manifest ../hack/manifest.yaml

# Apply
planton tofu apply --manifest ../hack/manifest.yaml --auto-approve

# Destroy
planton tofu destroy --manifest ../hack/manifest.yaml --auto-approve
```

### Standalone Terraform

```hcl
module "documentdb" {
  source = "./path/to/module"

  metadata = {
    name = "my-docdb"
    id   = "my-docdb-cluster"
    org  = "my-org"
    env  = "prod"
    labels = {
      key   = ""
      value = ""
    }
    annotations = {
      key   = ""
      value = ""
    }
    tags = []
  }

  spec = {
    subnets = [
      { value = "subnet-12345678" },
      { value = "subnet-87654321" }
    ]
    db_subnet_group = { value = "" }
    security_groups = []
    allowed_cidrs   = ["10.0.0.0/16"]
    vpc             = { value = "vpc-12345678" }
    engine_version  = "5.0.0"
    port            = 27017
    master_username = "docdbadmin"
    master_password = "MySecurePassword123!"
    instance_count  = 3
    instance_class  = "db.r6g.large"
    storage_encrypted = true
    kms_key           = { value = "" }
    backup_retention_period      = 7
    preferred_backup_window      = "03:00-04:00"
    preferred_maintenance_window = "sun:05:00-sun:06:00"
    deletion_protection          = true
    skip_final_snapshot          = false
    final_snapshot_identifier    = "my-docdb-final"
    enabled_cloudwatch_logs_exports = ["audit", "profiler"]
    apply_immediately               = false
    auto_minor_version_upgrade      = true
    cluster_parameter_group_name    = ""
    cluster_parameters              = []
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `cluster_endpoint` | Primary writer endpoint |
| `cluster_reader_endpoint` | Reader endpoint for read replicas |
| `cluster_id` | Cluster identifier |
| `cluster_arn` | Cluster ARN |
| `cluster_port` | Connection port |
| `db_subnet_group_name` | Subnet group name |
| `security_group_id` | Security group ID (if created) |
| `cluster_parameter_group_name` | Parameter group name (if created) |
| `connection_string` | MongoDB-compatible connection string template |
| `cluster_resource_id` | Internal AWS resource ID |
