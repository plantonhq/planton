# AlicloudDnsRecord Terraform Examples

## A Record

```hcl
module "dns_record" {
  source = "."

  metadata = {
    name = "web-server"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
    rr          = "www"
    type        = "A"
    value       = "203.0.113.10"
    ttl         = 600
  }
}
```

## CNAME Record

```hcl
module "dns_record" {
  source = "."

  metadata = {
    name = "cdn-alias"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
    rr          = "cdn"
    type        = "CNAME"
    value       = "example.com.cdn-provider.com"
  }
}
```

## MX Record with Priority

```hcl
module "dns_record" {
  source = "."

  metadata = {
    name = "mail-primary"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
    rr          = "@"
    type        = "MX"
    value       = "mx1.example.com"
    priority    = 5
    ttl         = 3600
  }
}
```

## Disabled Record

```hcl
module "dns_record" {
  source = "."

  metadata = {
    name = "staging-record"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
    rr          = "staging-api"
    type        = "A"
    value       = "10.0.1.100"
    status      = "DISABLE"
    remark      = "Pre-staged for next release"
  }
}
```
