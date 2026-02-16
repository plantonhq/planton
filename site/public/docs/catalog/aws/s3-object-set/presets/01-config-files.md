---
title: "Configuration Files"
description: "This preset uploads a set of application configuration files to an S3 bucket. It demonstrates the inline content pattern for managing JSON and YAML configuration files as infrastructure, with proper..."
type: "preset"
rank: "01"
presetSlug: "01-config-files"
componentSlug: "s3-object-set"
componentTitle: "S3 Object Set"
provider: "aws"
icon: "package"
order: 1
---

# Configuration Files

This preset uploads a set of application configuration files to an S3 bucket. It demonstrates the inline content pattern for managing JSON and YAML configuration files as infrastructure, with proper MIME content types set for each object. Replace the example content with your actual application configuration.

## When to Use

- Managing application configuration files alongside infrastructure
- Deploying environment-specific settings to S3 for applications that read config from S3
- Seed data or bootstrap configuration that needs to exist before the application starts

## Key Configuration Choices

- **Inline text content** (`content`) -- Configuration is declared directly in the manifest; changes are tracked as infrastructure changes
- **Proper content types** -- JSON and YAML files have correct MIME types set for downstream consumers
- **Organized key paths** (`config/...`) -- Objects are organized under a `config/` prefix for clean bucket structure

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<s3-bucket-name>` | Target S3 bucket name or ID | AWS S3 console or `AwsS3Bucket` status outputs |
| `<aws-region>` | AWS region where the bucket is located (e.g., `us-east-1`) | Must match the bucket's region |
| `<environment-name>` | Environment identifier (e.g., `production`, `staging`) | Your deployment configuration |
| `<database-host>` | Database hostname or endpoint | AWS RDS console or `AwsRdsInstance` status outputs |
| `<database-name>` | Database name | Your application's database configuration |
