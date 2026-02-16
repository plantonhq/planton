---
title: "Presets"
description: "Ready-to-deploy configuration presets for SageMaker Domain"
type: "preset-list"
componentSlug: "sagemaker-domain"
componentTitle: "SageMaker Domain"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-basic-jupyter-domain"
    rank: "01"
    title: "Preset: Basic JupyterLab Domain"
    excerpt: "A minimal SageMaker Domain for getting started with JupyterLab in your VPC."
  - slug: "02-production-vpc-only"
    rank: "02"
    title: "Preset: Production VPC-Only Domain"
    excerpt: "A security-hardened SageMaker Domain for production ML teams with SSO authentication, VPC-only networking, KMS encryption, and cost management via idle shutdown."
  - slug: "03-ml-team-with-custom-images"
    rank: "03"
    title: "Preset: ML Team with Custom Images"
    excerpt: "A fully-featured SageMaker Domain for advanced ML teams that need custom Docker images, GPU compute, Docker build capabilities, notebook sharing, and auto-cloned code repositories."
---

# SageMaker Domain Presets

Ready-to-deploy configuration presets for SageMaker Domain. Each preset is a complete manifest you can copy, customize, and deploy.
