# AliCloudRedisInstance Terraform Module

This Terraform module provisions an Alibaba Cloud Redis (KVStore) instance.

## Resources Created

- `alicloud_kvstore_instance` -- the Redis instance

## Local Development

```bash
cd apis/dev/planton/provider/alicloud/alicloudredisinstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Outputs

| Name | Description |
|---|---|
| `instance_id` | Redis instance ID |
| `connection_domain` | Intranet connection domain |
| `private_connection_port` | Private connection port |
| `private_ip` | Private IP address |
