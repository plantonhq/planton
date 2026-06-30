# AwsRedshiftCluster Terraform Module

This directory contains the Terraform module for provisioning an Amazon Redshift
data warehouse cluster with optional subnet group, security group, parameter
group, and audit logging.

## Usage

```hcl
module "redshift_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "my-warehouse"
  }

  spec = {
    node_type              = "ra3.xlplus"
    number_of_nodes        = 2
    database_name          = "warehouse"
    master_username        = "admin"
    manage_master_password = true
    subnet_ids             = ["subnet-aaa", "subnet-bbb"]
    encrypted              = true
    skip_final_snapshot    = false
    final_snapshot_identifier = "my-warehouse-final"
  }
}
```

## Inputs

| Name | Type | Required | Description |
|------|------|----------|-------------|
| metadata | object | yes | Planton resource metadata (name, org, env, id) |
| spec | object | yes | AwsRedshiftCluster specification |
| provider_config | object | no | AWS provider credentials and region |

## Outputs

| Name | Description |
|------|-------------|
| cluster_identifier | Unique identifier of the Redshift cluster |
| cluster_arn | ARN of the cluster |
| cluster_namespace_arn | Namespace ARN for data sharing |
| endpoint | Connection endpoint (address:port) |
| dns_name | DNS hostname (without port) |
| database_name | Default database name |
| port | TCP port for connections |
| subnet_group_name | Managed subnet group name (if created) |
| security_group_id | Managed security group ID (if created) |
| parameter_group_name | Managed parameter group name (if created) |
| master_password_secret_arn | Secrets Manager secret ARN (if managed password) |

## Conditional Resources

The module creates resources based on input:

- **Subnet group** — Created when `spec.subnet_ids` has ≥ 2 entries
- **Security group** — Created when `spec.security_group_ids` or `spec.allowed_cidr_blocks` are provided
- **Parameter group** — Created when `spec.parameters` has entries
- **Logging** — Created when `spec.logging` is provided

## Provider Requirements

- `hashicorp/aws` >= 5.0
