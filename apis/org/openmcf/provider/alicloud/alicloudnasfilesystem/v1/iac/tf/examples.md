# Terraform Examples

## Minimal Standard NFS

```hcl
module "nas" {
  source = "."

  metadata = {
    name = "dev-share"
  }

  spec = {
    region        = "cn-hangzhou"
    protocol_type = "NFS"
    storage_type  = "Performance"
    vpc_id        = "vpc-abc123"
    vswitch_id    = "vsw-abc123"
  }
}
```

## Production with Encryption and Custom Access Rules

```hcl
module "nas" {
  source = "."

  metadata = {
    name = "prod-storage"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region        = "cn-shanghai"
    protocol_type = "NFS"
    storage_type  = "Performance"
    description   = "Production shared storage"

    encryption = {
      encrypt_type = 1
    }

    vpc_id     = "vpc-prod-001"
    vswitch_id = "vsw-prod-001"

    access_rules = [
      {
        source_cidr_ip   = "10.0.1.0/24"
        rw_access_type   = "RDWR"
        user_access_type = "root_squash"
      }
    ]

    tags = {
      team = "platform"
    }
  }
}
```

## Extreme NAS with KMS Encryption

```hcl
module "nas" {
  source = "."

  metadata = {
    name = "hpc-scratch"
  }

  spec = {
    region           = "cn-hangzhou"
    file_system_type = "extreme"
    protocol_type    = "NFS"
    storage_type     = "advance"
    capacity         = 500
    zone_id          = "cn-hangzhou-a"

    encryption = {
      encrypt_type = 2
      kms_key_id   = "cmk-abc123"
    }

    vpc_id     = "vpc-hpc-001"
    vswitch_id = "vsw-hpc-001"

    access_rules = [
      {
        source_cidr_ip = "10.0.0.0/8"
      }
    ]
  }
}
```
