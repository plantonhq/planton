---
title: "Presets"
description: "Ready-to-deploy configuration presets for Deployment"
type: "preset-list"
componentSlug: "deployment"
componentTitle: "Deployment"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-web-service"
    rank: "01"
    title: "Web Service Deployment"
    excerpt: "This preset deploys a single-replica web application with an HTTP port and ingress. It is the most common Kubernetes Deployment pattern: a containerized web service exposed via an ingress hostname."
  - slug: "02-web-service-with-hpa"
    rank: "02"
    title: "Production Web Service with HPA"
    excerpt: "This preset deploys a production-grade web application with horizontal pod autoscaling, a pod disruption budget, and a zero-downtime rolling update strategy."
  - slug: "03-worker"
    rank: "03"
    title: "Background Worker Deployment"
    excerpt: "This preset deploys a background worker process without ingress. Use this for queue consumers, event processors, or any long-running process that does not serve HTTP traffic."
---

# Deployment Presets

Ready-to-deploy configuration presets for Deployment. Each preset is a complete manifest you can copy, customize, and deploy.
