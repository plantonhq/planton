# AlicloudNatGateway Terraform Module

This Terraform module provisions an Alibaba Cloud Enhanced NAT Gateway with EIP association and SNAT entries.

## Resources Created

- `alicloud_nat_gateway` -- the NAT Gateway
- `alicloud_eip_association` -- binds an EIP to the NAT Gateway
- `alicloud_snat_entry` -- one per SNAT entry, mapping private traffic to the EIP
- `data.alicloud_eip_addresses` -- looks up the EIP's public IP from its allocation ID

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudnatgateway/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Outputs

| Name | Description |
| --- | --- |
| `nat_gateway_id` | NAT Gateway resource ID |
| `nat_gateway_name` | NAT Gateway name |
| `snat_table_id` | SNAT table ID |
| `forward_table_id` | Forward (DNAT) table ID |
