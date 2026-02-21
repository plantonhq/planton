# Terraform Module to Deploy AliCloudStorageBucket

This module provisions an Alibaba Cloud OSS bucket with configurable access control, storage class, redundancy, versioning, server-side encryption, lifecycle rules, CORS configuration, and access logging.

Generated from the proto schema for `AliCloudStorageBucket`.

## Usage

```hcl
module "oss_bucket" {
  source = "./path/to/module"

  metadata = {
    name = "my-bucket"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region      = "cn-hangzhou"
    bucket_name = "my-app-bucket"

    versioning_enabled = true

    server_side_encryption = {
      sse_algorithm = "AES256"
    }

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
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `bucket_name` | The bucket name (also the bucket ID) |
| `extranet_endpoint` | Public internet endpoint |
| `intranet_endpoint` | VPC-internal endpoint |

## Further Reading

- [examples.md](./examples.md) -- YAML and HCL usage examples
