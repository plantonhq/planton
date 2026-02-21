# AlicloudStorageBucket Terraform Examples

## Minimal Bucket

```hcl
module "dev_bucket" {
  source = "."

  metadata = {
    name = "dev-bucket"
  }

  spec = {
    region      = "cn-hangzhou"
    bucket_name = "dev-assets-bucket"
  }
}
```

## Production Bucket with Versioning and Encryption

```hcl
module "prod_bucket" {
  source = "."

  metadata = {
    name = "prod-bucket"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region          = "cn-shanghai"
    bucket_name     = "prod-platform-data"
    redundancy_type = "ZRS"

    versioning_enabled = true

    server_side_encryption = {
      sse_algorithm = "AES256"
    }

    tags = {
      team       = "platform"
      costCenter = "engineering"
    }
  }
}
```

## Archive Bucket with Lifecycle Rules

```hcl
module "archive_bucket" {
  source = "."

  metadata = {
    name = "log-archive"
    env  = "production"
  }

  spec = {
    region      = "cn-hangzhou"
    bucket_name = "platform-log-archive"

    versioning_enabled = true

    lifecycle_rules = [
      {
        prefix          = ""
        enabled         = true
        expiration_days = 365
        transitions = [
          { days = 30, storage_class = "IA" },
          { days = 90, storage_class = "Archive" }
        ]
        abort_multipart_upload_days        = 7
        noncurrent_version_expiration_days = 30
      }
    ]

    tags = {
      purpose   = "log-archive"
      retention = "1-year"
    }
  }
}
```
