# AliCloudNatGateway Pulumi Module

This Pulumi module provisions an Alibaba Cloud Enhanced NAT Gateway with EIP association and SNAT entries.

## Resources Created

- `alicloud:vpc/natGateway:NatGateway` -- the NAT Gateway
- `alicloud:ecs/eipAssociation:EipAssociation` -- binds an EIP to the NAT Gateway
- `alicloud:vpc/snatEntry:SnatEntry` -- one per SNAT entry, mapping private traffic to the EIP

## Architecture

The module resolves the EIP's public IP address internally using the `ecs.GetEipAddresses` data source, so the user only needs to provide the EIP allocation ID. The SNAT entries are parented to the NAT Gateway for clean resource hierarchy.

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudnatgateway/v1/iac/pulumi
go build ./...
go vet ./...
```

## Stack Outputs

| Name | Description |
| --- | --- |
| `nat_gateway_id` | NAT Gateway resource ID |
| `nat_gateway_name` | NAT Gateway name |
| `snat_table_id` | SNAT table ID |
| `forward_table_id` | Forward (DNAT) table ID |
