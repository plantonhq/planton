---
title: "Presets"
description: "Ready-to-deploy configuration presets for ECS Service"
type: "preset-list"
componentSlug: "ecs-service"
componentTitle: "ECS Service"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-web-service-alb"
    rank: "01"
    title: "Web Service with ALB"
    excerpt: "This preset deploys a Fargate-based ECS service fronted by an Application Load Balancer using path-based routing. It runs 2 replicas across two Availability Zones with CloudWatch logging and a..."
  - slug: "02-api-service-autoscaling"
    rank: "02"
    title: "API Service with Autoscaling"
    excerpt: "This preset deploys a Fargate-based ECS service with hostname-based ALB routing and CPU-based autoscaling. It starts with 2 replicas and scales to 10 based on a 75% CPU target. This is the standard..."
  - slug: "03-background-worker"
    rank: "03"
    title: "Background Worker"
    excerpt: "This preset deploys a Fargate-based ECS service for background processing without an ALB. The container has no exposed port -- it pulls work from a queue (SQS, Redis, etc.) or runs scheduled tasks...."
---

# ECS Service Presets

Ready-to-deploy configuration presets for ECS Service. Each preset is a complete manifest you can copy, customize, and deploy.
