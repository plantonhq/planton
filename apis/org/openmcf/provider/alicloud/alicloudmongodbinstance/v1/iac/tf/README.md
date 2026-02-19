# AlicloudMongodbInstance Terraform Module

This Terraform module provisions an Alibaba Cloud ApsaraDB for MongoDB replica-set instance.

## Resources Created

- `alicloud_mongodb_instance` -- the MongoDB replica-set instance

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudmongodbinstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Outputs

| Name | Description |
|---|---|
| `instance_id` | MongoDB instance ID |
| `replica_set_name` | Replica set name for connection strings |
