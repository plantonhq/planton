---
title: "Preset: ML Team with Custom Images"
description: "A fully-featured SageMaker Domain for advanced ML teams that need custom Docker images, GPU compute, Docker build capabilities, notebook sharing, and auto-cloned code repositories."
type: "preset"
rank: "03"
presetSlug: "03-ml-team-with-custom-images"
componentSlug: "sagemaker-domain"
componentTitle: "SageMaker Domain"
provider: "aws"
icon: "package"
order: 3
---

# Preset: ML Team with Custom Images

A fully-featured SageMaker Domain for advanced ML teams that need custom Docker images,
GPU compute, Docker build capabilities, notebook sharing, and auto-cloned code repositories.

## When to Use

- ML platform teams with custom training frameworks
- Teams building custom Docker images for training and inference
- GPU-intensive workloads (computer vision, NLP, deep learning)
- Collaborative teams that need notebook output sharing

## Configuration Highlights

- **Auth mode**: SSO (enterprise identity management)
- **Network**: VpcOnly (secure, Docker pulls restricted to trusted accounts)
- **Encryption**: Customer-managed KMS for EFS and shared notebook outputs
- **Docker**: Enabled with trusted account restrictions
- **JupyterLab**: `ml.m5.large` default, 3-hour idle timeout, 2 auto-cloned repos
- **Custom images**: PyTorch GPU and TensorFlow custom images for JupyterLab
- **KernelGateway**: `ml.g4dn.xlarge` (GPU) with custom ML framework image
- **Sharing**: Notebook outputs persisted to S3 with KMS encryption
- **Storage**: 50 GB default / 500 GB max EBS per space

## Cost Estimate

Per-user compute (with 3-hour idle timeout, 8-hour workday):
- JupyterLab `ml.m5.large`: ~$0.77/day per user (~$23/month)
- KernelGateway `ml.g4dn.xlarge` (GPU): ~$4.20/day per user when active
- EBS storage: $0.10/GB-month (50-500 GB per space)
- EFS: $0.30/GB-month for home directories
- S3: $0.023/GB-month for shared notebook outputs

## Customization

- Adjust `idleTimeoutInMinutes` based on team workflow (shorter = lower cost)
- Add more `customImages` as new ML frameworks are onboarded
- Increase `maximumEbsVolumeSizeInGb` for large dataset workflows
- Add additional `codeRepositories` for project-specific repos
