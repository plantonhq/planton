# DigitalOcean DNS Record - Terraform Module

This Terraform module provisions a DigitalOcean DNS record.

## Usage

### Via Planton CLI

```bash
planton tofu apply -f manifest.yaml
```

### Direct Terraform Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "www-record"
  }

  spec = {
    domain      = "example.com"
    name        = "www"
    type        = "A"
    value       = "192.0.2.1"
    ttl_seconds = 3600
  }

  digitalocean_token = var.digitalocean_token
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata | object | yes |
| `spec` | DNS record specification | object | yes |
| `digitalocean_token` | DigitalOcean API token | string | yes |

### Spec Object

| Field | Description | Type | Default |
|-------|-------------|------|---------|
| `domain` | Domain name (DNS zone) | string | - |
| `name` | Record name (@ for root) | string | - |
| `type` | Record type (A, AAAA, CNAME, etc.) | string | - |
| `value` | Record value | string | - |
| `ttl_seconds` | Time to live in seconds | number | 1800 |
| `priority` | Priority for MX/SRV | number | 0 |
| `weight` | Weight for SRV | number | 0 |
| `port` | Port for SRV | number | 0 |
| `flags` | Flags for CAA | number | 0 |
| `tag` | Tag for CAA | string | "" |

## Outputs

| Name | Description |
|------|-------------|
| `record_id` | The unique ID of the created DNS record |
| `hostname` | The fully qualified hostname |
| `record_type` | The type of DNS record created |
| `domain` | The domain where the record was created |
| `ttl_seconds` | The TTL applied to the record |

## Examples

### A Record

```hcl
spec = {
  domain = "example.com"
  name   = "www"
  type   = "A"
  value  = "192.0.2.1"
}
```

### MX Record

```hcl
spec = {
  domain   = "example.com"
  name     = "@"
  type     = "MX"
  value    = "mail.example.com"
  priority = 10
}
```

### SRV Record

```hcl
spec = {
  domain   = "example.com"
  name     = "_sip._tcp"
  type     = "SRV"
  value    = "sipserver.example.com"
  priority = 10
  weight   = 5
  port     = 5060
}
```

### CAA Record

```hcl
spec = {
  domain = "example.com"
  name   = "@"
  type   = "CAA"
  value  = "letsencrypt.org"
  flags  = 0
  tag    = "issue"
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| digitalocean | ~> 2.0 |
