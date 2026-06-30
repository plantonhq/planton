# AliCloudVpnGateway Terraform Module

This Terraform module provisions an Alibaba Cloud VPN Gateway with customer gateways and IPsec VPN connections.

## Resources Created

- `alicloud_vpn_gateway` -- the VPN Gateway
- `alicloud_vpn_customer_gateway` -- one per connection (via `for_each`)
- `alicloud_vpn_connection` -- one per connection (via `for_each`)

## Files

| File | Purpose |
|------|---------|
| `main.tf` | VPN Gateway resource |
| `connections.tf` | Customer gateways and VPN connections via `for_each` |
| `variables.tf` | Input variables (from proto spec) |
| `outputs.tf` | Output values (matching stack_outputs.proto) |
| `locals.tf` | Computed values, tags, connection map |
| `provider.tf` | Provider configuration |

## Local Development

```bash
cd apis/dev/planton/provider/alicloud/alicloudvpngateway/v1/iac/tf
terraform init -backend=false
terraform validate
```
