# AlicloudRedisInstance Pulumi Module

This Pulumi module provisions an Alibaba Cloud Redis (KVStore) instance.

## Resources Created

- `alicloud:kvstore/instance:Instance` -- the Redis instance

## Architecture

The module creates the instance with the specified configuration, applying tags, networking, security, and operational settings. Unlike the RDS module, no sub-resources (databases, accounts) are bundled -- Redis authentication is handled by the instance-level password.

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudredisinstance/v1/iac/pulumi
go build ./...
go vet ./...
```

## Stack Outputs

| Name | Description |
|---|---|
| `instance_id` | Redis instance ID |
| `connection_domain` | Intranet connection domain |
| `private_connection_port` | Private connection port |
| `private_ip` | Private IP address |
