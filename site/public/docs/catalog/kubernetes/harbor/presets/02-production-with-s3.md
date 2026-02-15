---
title: "Production Harbor with S3 Storage"
description: "This preset deploys Harbor with S3-compatible storage for container image layers. Provides durable, scalable storage independent of the Kubernetes cluster's local disks."
type: "preset"
rank: "02"
presetSlug: "02-production-with-s3"
componentSlug: "harbor"
componentTitle: "Harbor"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production Harbor with S3 Storage

This preset deploys Harbor with S3-compatible storage for container image layers. Provides durable, scalable storage independent of the Kubernetes cluster's local disks.

## When to Use

- Production container registries where image storage must be durable and scalable
- Multi-cluster environments where registry storage should be independent of any single cluster
- AWS, MinIO, or any S3-compatible object storage backend

## Key Configuration Choices

- **S3 storage backend** -- container images stored in S3; eliminates local disk dependency
- **Explicit container resources** -- production-appropriate for core and registry components
- **Ingress enabled** -- exposes Harbor at the specified hostname

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-harbor.example.com>` | Hostname for the Harbor registry and UI | Your DNS provider |
| `<your-aws-region>` | S3 bucket region (e.g., `us-east-1`) | AWS Console or your S3-compatible provider |
| `<your-s3-bucket-name>` | S3 bucket for storing container image layers | AWS Console > S3 or MinIO console |
| `<your-s3-access-key>` | S3 access key | AWS IAM or MinIO admin |
| `<your-s3-secret-key>` | S3 secret key | AWS IAM or MinIO admin |

## Related Presets

- **01-minimal** -- Default filesystem storage for development
