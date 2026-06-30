# AliCloudMongodbInstance Pulumi Module

This Pulumi module provisions an Alibaba Cloud ApsaraDB for MongoDB replica-set instance.

## Resources Created

- `alicloud:mongodb/instance:Instance` -- the MongoDB replica-set instance

## Architecture

The module creates a single MongoDB replica-set instance with configurable replication factor, storage engine, multi-zone HA, and optional encryption (TDE or cloud disk). No sub-resources (accounts, databases) are bundled -- MongoDB authentication is handled by the instance-level `accountPassword`.

## Local Development

```bash
cd apis/dev/planton/provider/alicloud/alicloudmongodbinstance/v1/iac/pulumi
go build ./...
go vet ./...
```

## Stack Outputs

| Name | Description |
|---|---|
| `instance_id` | MongoDB instance ID |
| `replica_set_name` | Replica set name for connection strings |
