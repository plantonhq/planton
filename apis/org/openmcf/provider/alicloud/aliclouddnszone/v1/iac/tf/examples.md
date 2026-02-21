# AlicloudDnsZone Terraform Examples

## Minimal Domain

```hcl
module "dns_domain" {
  source = "."

  metadata = {
    name = "my-domain"
  }

  spec = {
    region      = "cn-hangzhou"
    domain_name = "example.com"
  }
}
```

## Domain with Tags and Resource Group

```hcl
module "dns_domain" {
  source = "."

  metadata = {
    name = "platform-domain"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region            = "cn-shanghai"
    domain_name       = "platform.example.com"
    remark            = "Primary platform domain"
    resource_group_id = "rg-prod-123"
    tags = {
      team       = "platform"
      costCenter = "engineering"
    }
  }
}
```

## Domain with Group Assignment

```hcl
module "dns_domain" {
  source = "."

  metadata = {
    name = "grouped-domain"
  }

  spec = {
    region      = "ap-southeast-1"
    domain_name = "services.example.com"
    group_id    = "group-abc123"
    remark      = "Microservices DNS zone"
  }
}
```
