# AliCloudRdsInstance Terraform Module

This Terraform module provisions an Alibaba Cloud RDS instance with bundled databases, accounts, and account privileges.

## Resources Created

- `alicloud_db_instance` -- the RDS instance
- `alicloud_db_database` -- one per database entry (via `for_each`)
- `alicloud_rds_account` -- one per account entry (via `for_each`)
- `alicloud_db_account_privilege` -- one per privilege entry (via `for_each`)

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Outputs

| Name | Description |
| --- | --- |
| `instance_id` | RDS instance ID |
| `connection_string` | Intranet connection endpoint |
| `port` | Database service port |
| `database_ids` | Map of database names to IDs |
