# Terraform Examples

## Minimal Zone

```hcl
module "private_zone" {
  source = "./path/to/module"

  metadata = {
    name = "minimal-zone"
  }

  spec = {
    region    = "cn-hangzhou"
    zone_name = "internal.example.com"
    vpc_attachments = [
      { vpc_id = "vpc-abc123" }
    ]
  }
}
```

## Zone with Records

```hcl
module "private_zone" {
  source = "./path/to/module"

  metadata = {
    name = "svc-zone"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region    = "cn-shanghai"
    zone_name = "svc.internal"
    vpc_attachments = [
      { vpc_id = "vpc-prod-app" }
    ]
    records = [
      { rr = "api", type = "A", value = "10.0.1.50", ttl = 120 },
      { rr = "db",  type = "A", value = "10.0.2.100" },
    ]
    tags = {
      team = "platform"
    }
  }
}
```

## Multi-VPC Cross-Region

```hcl
module "private_zone" {
  source = "./path/to/module"

  metadata = {
    name = "shared-zone"
  }

  spec = {
    region    = "cn-hangzhou"
    zone_name = "shared.corp"
    vpc_attachments = [
      { vpc_id = "vpc-hz" },
      { vpc_id = "vpc-sh", region_id = "cn-shanghai" },
    ]
  }
}
```
