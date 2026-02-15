---
title: "Presets"
description: "Ready-to-deploy configuration presets for ECS Cluster"
type: "preset-list"
componentSlug: "ecs-cluster"
componentTitle: "ECS Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-fargate-standard"
    rank: "01"
    title: "Standard Fargate Cluster"
    excerpt: "This preset creates an ECS cluster using AWS Fargate with CloudWatch Container Insights enabled. Fargate eliminates the need to manage EC2 instances for container workloads -- AWS handles the compute..."
  - slug: "02-fargate-cost-optimized"
    rank: "02"
    title: "Fargate Cost-Optimized Cluster"
    excerpt: "This preset creates an ECS cluster with both Fargate and Fargate Spot capacity providers, using a weighted strategy that runs approximately 80% of scaled tasks on Spot for significant cost savings..."
---

# ECS Cluster Presets

Ready-to-deploy configuration presets for ECS Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
