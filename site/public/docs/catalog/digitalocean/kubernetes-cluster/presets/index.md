---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kubernetes Cluster"
type: "preset-list"
componentSlug: "kubernetes-cluster"
componentTitle: "Kubernetes Cluster"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-production-ha"
    rank: "01"
    title: "Production HA Kubernetes Cluster"
    excerpt: "This preset creates a production-grade DigitalOcean Kubernetes (DOKS) cluster with a highly available control plane, autoscaling default node pool, automatic patch upgrades, a scheduled maintenance..."
  - slug: "02-development"
    rank: "02"
    title: "Development Kubernetes Cluster"
    excerpt: "This preset creates a minimal DigitalOcean Kubernetes cluster for development and testing. It uses a non-HA control plane, a fixed-size node pool with smaller instances, and no API server firewall --..."
---

# Kubernetes Cluster Presets

Ready-to-deploy configuration presets for Kubernetes Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
