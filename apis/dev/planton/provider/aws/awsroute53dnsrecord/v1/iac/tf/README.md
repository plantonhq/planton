# AWS Route53 DNS Record Terraform Module

This Terraform module creates DNS records in AWS Route53 hosted zones.

## Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "www-example"
  }

  spec = {
    zone_id = {
      value = "Z1234567890ABC"
    }
    name   = "www.example.com"
    type   = "A"
    ttl    = 300
    values = ["192.0.2.1"]
  }
}
```

### Alias Record to ALB

```hcl
module "alb_alias" {
  source = "./path/to/module"

  metadata = {
    name = "api-alb"
  }

  spec = {
    zone_id = {
      value = "Z1234567890ABC"
    }
    name = "api.example.com"
    type = "A"
    alias_target = {
      dns_name = {
        value = "my-alb-1234567890.us-east-1.elb.amazonaws.com"
      }
      zone_id = {
        value = "Z35SXDOTRQ7X7K"
      }
      evaluate_target_health = true
    }
  }
}
```

## Variables

### metadata

Resource metadata including the name.

### spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | object | Yes | Route53 zone ID (StringValueOrRef structure) |
| `name` | string | Yes | DNS record name |
| `type` | string | Yes | Record type (A, AAAA, CNAME, MX, TXT, etc.) |
| `ttl` | number | No | TTL in seconds (default: 300, ignored for alias) |
| `values` | list(string) | No | Record values (for standard records) |
| `alias_target` | object | No | Alias target configuration |
| `routing_policy` | object | No | Advanced routing configuration |
| `health_check_id` | string | No | Health check for failover |
| `set_identifier` | string | No | Identifier for routing policies |

## Outputs

| Output | Description |
|--------|-------------|
| `fqdn` | Fully qualified domain name |
| `record_type` | DNS record type |
| `zone_id` | Route53 hosted zone ID |
| `is_alias` | Whether this is an alias record |
| `set_identifier` | Routing policy set identifier |

## StringValueOrRef Structure

Fields like `zone_id` and `alias_target.dns_name` use a StringValueOrRef structure:

```hcl
# Literal value
zone_id = {
  value = "Z1234567890ABC"
}

# The CLI resolves value_from references before passing to Terraform
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | >= 5.0 |
