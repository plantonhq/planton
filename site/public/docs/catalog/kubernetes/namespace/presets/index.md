---
title: "Presets"
description: "Ready-to-deploy configuration presets for Namespace"
type: "preset-list"
componentSlug: "namespace"
componentTitle: "Namespace"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Namespace"
    excerpt: "This preset creates a Kubernetes namespace with a small built-in resource profile and baseline pod security. Suitable for most development and staging workloads where basic resource guardrails and..."
  - slug: "02-production-with-quotas"
    rank: "02"
    title: "Production Namespace with Custom Quotas"
    excerpt: "This preset creates a hardened production namespace with custom resource quotas, default container limits, network isolation, and restricted pod security. Designed for production workloads where..."
  - slug: "03-istio-enabled"
    rank: "03"
    title: "Istio-Enabled Namespace"
    excerpt: "This preset creates a namespace with Istio service mesh sidecar injection enabled. All pods deployed in this namespace will automatically receive an Istio sidecar proxy for mTLS, traffic management,..."
---

# Namespace Presets

Ready-to-deploy configuration presets for Namespace. Each preset is a complete manifest you can copy, customize, and deploy.
