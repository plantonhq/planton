---
title: "Presets"
description: "Ready-to-deploy configuration presets for SaeApplication"
type: "preset-list"
componentSlug: "saeapplication"
componentTitle: "SaeApplication"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-container-image-production"
    rank: "01"
    title: "Production Container Image Application"
    excerpt: "This preset creates a production SAE application deployed as a container image inside a VPC. Three replicas provide horizontal redundancy, with liveness and readiness HTTP probes ensuring traffic is..."
  - slug: "02-java-fatjar-production"
    rank: "02"
    title: "Production Java FatJar Application"
    excerpt: "This preset creates a production SAE application deployed as a Java FatJar package with JVM tuning, Spring Boot Actuator health checks, and VPC connectivity. Three replicas with rolling updates..."
  - slug: "03-container-image-development"
    rank: "03"
    title: "Development Container Image Application"
    excerpt: "This preset creates a minimal SAE application for development and testing. A single replica with the smallest compute tier (0.5 vCPU, 1 GB) keeps costs low. SAE-managed networking is used instead of..."
---

# SaeApplication Presets

Ready-to-deploy configuration presets for SaeApplication. Each preset is a complete manifest you can copy, customize, and deploy.
