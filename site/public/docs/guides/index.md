---
title: "Guides"
description: "How-to guides for Planton — manifests, credentials, provider setup, state backends, Kustomize, CI/CD, and migration"
icon: "guide"
order: 30
---

# Guides

Practical how-to guides for deploying and managing infrastructure with Planton. Each guide focuses on a specific task and assumes you have completed the [Getting Started](/docs/getting-started) walkthrough.

## Core Guides

- **[Writing Manifests](./manifests)** — Find the right component, write a manifest, validate it, and deploy
- **[Credentials](./credentials)** — How Planton loads credentials, with a quick reference for all 17 providers
- **[Kustomize Integration](./kustomize)** — Manage multi-environment deployments with base manifests and overlays
- **[State Backends](./state-backends)** — Configure state storage for Pulumi, OpenTofu, and Terraform
- **[Advanced Usage](./advanced-usage)** — Runtime overrides with `--set`, URL manifests, module customization, and scripting

## Provider Setup

Detailed credential and IAM configuration for the three major cloud providers:

- **[AWS Provider Setup](./aws-provider-setup)** — IAM users, roles, environment variables, and `-p` config files
- **[GCP Provider Setup](./gcp-provider-setup)** — Service accounts, Application Default Credentials, and Workload Identity
- **[Azure Provider Setup](./azure-provider-setup)** — Service principals, RBAC roles, and Azure CLI authentication

## Workflows

- **[CI/CD Integration](./cicd-integration)** — GitHub Actions, GitLab CI, non-interactive flags, and credential injection
- **[Migrating to Planton](./migrating-to-planton)** — Step-by-step migration from raw Terraform or Pulumi
