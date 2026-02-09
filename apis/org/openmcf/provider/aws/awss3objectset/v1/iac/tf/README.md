# Terraform Module to Deploy AwsS3ObjectSet

## Overview

This Terraform module uploads one or more S3 objects to a target bucket based on the `AwsS3ObjectSet` API resource specification. It uses `for_each` over the objects list to create individual `aws_s3_object` resources with proper tag inheritance.

## Usage

```hcl
module "s3_objects" {
  source = "./path/to/module"

  metadata = {
    name        = "app-config-objects"
    id          = "s3objs-abc123"
    org         = "my-org"
    env         = "production"
    labels      = {}
    annotations = {}
    tags        = []
  }

  spec = {
    bucket     = "my-app-bucket"
    aws_region = "us-east-1"
    tags = {
      environment = "production"
    }
    objects = [
      {
        key          = "config/app.json"
        content      = jsonencode({ database = "postgres", port = 5432 })
        content_type = "application/json"
      },
      {
        key          = "index.html"
        content      = "<html><body><h1>Hello</h1></body></html>"
        content_type = "text/html"
        cache_control = "max-age=300"
      }
    ]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata (name, id, org, env, labels) | object | yes |
| spec | Resource specification (bucket, aws_region, objects) | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| object_etags | Map of object key to ETag |
| object_version_ids | Map of object key to version ID |

## Resources Created

- `aws_s3_object` - One per object in the spec, keyed by S3 object key
