# AliCloudRdsInstance Pulumi Module

This Pulumi module provisions an Alibaba Cloud RDS instance with bundled databases, accounts, and account privileges.

## Resources Created

- `alicloud:rds/instance:Instance` -- the RDS instance
- `alicloud:rds/database:Database` -- one per database entry
- `alicloud:rds/rdsAccount:RdsAccount` -- one per account entry
- `alicloud:rds/accountPrivilege:AccountPrivilege` -- one per privilege entry within each account

## Architecture

The module creates the instance first, then databases (parented to the instance), then accounts (parented to the instance), and finally grants account privileges (parented to the account). Engine-specific defaults (character sets) are applied automatically when not explicitly set.

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1/iac/pulumi
go build ./...
go vet ./...
```

## Stack Outputs

| Name | Description |
| --- | --- |
| `instance_id` | RDS instance ID |
| `connection_string` | Intranet connection endpoint |
| `port` | Database service port |
| `database_ids` | Map of database names to IDs |
