# Terraform Module to Deploy AlicloudDnsZone

This module provisions an Alibaba Cloud DNS domain in the Alidns service with optional group assignment, resource group placement, remarks, and automatic tag management.

## Usage

```hcl
module "dns_domain" {
  source = "./path/to/module"

  metadata = {
    name = "my-domain"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
    remark      = "Primary platform domain"
    tags = {
      team = "platform"
    }
  }
}
```

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, org, env, labels) |
| `spec` | DNS domain specification (region, domain_name, group_id, remark, resource_group_id, tags) |

## Outputs

| Output | Description |
|--------|-------------|
| `domain_id` | The domain ID assigned by Alibaba Cloud |
| `domain_name` | The domain name as registered |
| `dns_servers` | DNS server names assigned by Alibaba Cloud |
| `group_name` | Computed domain group name |
| `puny_code` | Punycode representation for internationalized domain names |
