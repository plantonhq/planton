---
title: "Presets"
description: "Ready-to-deploy configuration presets for Batch Compute Environment"
type: "preset-list"
componentSlug: "batch-compute-environment"
componentTitle: "Batch Compute Environment"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-fargate-batch"
    rank: "01"
    title: "Fargate Batch (Serverless)"
    excerpt: "This preset creates a serverless AWS Batch compute environment using Fargate. AWS manages all infrastructure — no EC2 instances to configure, patch, or scale. Ideal for teams that want to run batch..."
  - slug: "02-ec2-managed-batch"
    rank: "02"
    title: "EC2 Managed Batch"
    excerpt: "This preset creates an EC2-based AWS Batch compute environment with auto-scaling from 0 to 512 vCPUs. Uses `optimal` instance type selection and two priority-separated job queues. Ideal for..."
  - slug: "03-spot-cost-optimized-batch"
    rank: "03"
    title: "Spot Cost-Optimized Batch"
    excerpt: "This preset creates a high-capacity Spot-based AWS Batch compute environment optimized for cost. Uses capacity-optimized allocation across multiple instance families with a fair-share scheduling..."
---

# Batch Compute Environment Presets

Ready-to-deploy configuration presets for Batch Compute Environment. Each preset is a complete manifest you can copy, customize, and deploy.
