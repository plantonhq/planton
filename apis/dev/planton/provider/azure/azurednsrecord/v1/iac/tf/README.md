# Azure DNS Record Terraform Module

This Terraform module creates DNS records in an existing Azure DNS Zone.

## Prerequisites

- Terraform 1.0+
- Azure credentials configured

## Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "www-record"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    resource_group = "my-dns-rg"
    zone_name      = "example.com"
    type           = "A"
    name           = "www"
    values         = ["192.0.2.1"]
    ttl_seconds    = 300
  }
}
```

## Variables

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env, labels) | object | yes |
| `spec.resource_group` | Azure resource group containing the DNS zone | string | yes |
| `spec.zone_name` | DNS zone name | string | yes |
| `spec.type` | DNS record type (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA, PTR) | string | yes |
| `spec.name` | Record name (use @ for apex) | string | yes |
| `spec.values` | Record values | list(string) | yes |
| `spec.ttl_seconds` | TTL in seconds | number | no (default: 300) |
| `spec.mx_priority` | MX record priority | number | no (default: 10) |

## Outputs

| Name | Description |
|------|-------------|
| `record_id` | Azure Resource Manager ID of the DNS record |
| `fqdn` | Fully qualified domain name |

## Supported Record Types

| Type | Values Format |
|------|---------------|
| A | IPv4 addresses |
| AAAA | IPv6 addresses |
| CNAME | Target hostname (single value) |
| MX | Mail server hostnames |
| TXT | Text values |
| NS | Nameserver hostnames |
| SRV | Service targets |
| CAA | CA authorization values |
| PTR | Reverse DNS targets |

## Examples

### A Record
```hcl
spec = {
  resource_group = "dns-rg"
  zone_name      = "example.com"
  type           = "A"
  name           = "www"
  values         = ["192.0.2.1", "192.0.2.2"]
}
```

### MX Record
```hcl
spec = {
  resource_group = "dns-rg"
  zone_name      = "example.com"
  type           = "MX"
  name           = "@"
  values         = ["mail1.example.com", "mail2.example.com"]
  mx_priority    = 10
}
```

### TXT Record (SPF)
```hcl
spec = {
  resource_group = "dns-rg"
  zone_name      = "example.com"
  type           = "TXT"
  name           = "@"
  values         = ["v=spf1 include:_spf.google.com ~all"]
}
```
