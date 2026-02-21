# Terraform Module to Deploy AlicloudDnsRecord

This module creates an Alibaba Cloud DNS record in the Alidns service. It supports all standard record types with configurable TTL, priority, resolution lines, and record status.

## Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "my-record"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
    rr          = "www"
    type        = "A"
    value       = "203.0.113.10"
  }
}
```

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, org, env, labels) |
| `spec` | DNS record specification (region, domain_name, rr, type, value, ttl, priority, line, status, remark) |

## Outputs

| Output | Description |
|--------|-------------|
| `record_id` | The record ID assigned by Alibaba Cloud |
