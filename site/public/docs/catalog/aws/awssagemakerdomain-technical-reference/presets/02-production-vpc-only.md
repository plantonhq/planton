---
title: "Preset: Production VPC-Only Domain"
description: "A security-hardened SageMaker Domain for production ML teams with SSO authentication, VPC-only networking, KMS encryption, and cost management via idle shutdown."
type: "preset"
rank: "02"
presetSlug: "02-production-vpc-only"
componentSlug: "awssagemakerdomain-technical-reference"
componentTitle: "AwsSagemakerDomain — Technical Reference"
provider: "aws"
icon: "package"
order: 2
---

# Preset: Production VPC-Only Domain

A security-hardened SageMaker Domain for production ML teams with SSO authentication,
VPC-only networking, KMS encryption, and cost management via idle shutdown.

## When to Use

- Production ML platforms with compliance requirements
- Enterprise teams with centralized identity via AWS IAM Identity Center
- Environments where data exfiltration prevention is mandatory
- Teams that need cost guardrails for compute resources

## Configuration Highlights

- **Auth mode**: SSO (centralized identity management via IAM Identity Center)
- **Network**: VpcOnly (all traffic stays within VPC, requires NAT for internet)
- **Encryption**: Customer-managed KMS key for EFS home directories
- **Security**: Domain-level and user-level security groups for layered isolation
- **IDE**: JupyterLab with `ml.t3.medium` default instance
- **Cost control**: 2-hour idle timeout (saves ~70% on compute vs always-on)
- **Storage**: 20 GB default / 200 GB max EBS per space
- **Landing page**: JupyterLab opens by default

## Cost Estimate

Domain infrastructure: ~$0.30/GB-month for EFS storage.
Per-user compute (with 2-hour idle timeout, 8-hour workday):
- `ml.t3.medium`: ~$0.40/day per user (~$12/month)
- EBS storage: $0.10/GB-month

## Customization

- Add `sharingSettings` to enable notebook output sharing to S3
- Add `dockerSettings` to enable custom container workflows
- Add `kernelGatewayAppSettings` for custom ML framework images
- Add `jupyterLabAppSettings.codeRepositories` for auto-cloned Git repos
