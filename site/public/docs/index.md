---
title: "Welcome"
description: "Planton documentation — infrastructure provisioning, application delivery, and cloud operations in one platform."
icon: welcome
order: 0
tags:
  - Overview
  - Introduction
  - Welcome
---

# Planton Documentation

Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, verifies the cost and permissions before anything is created, and publishes it as templates your whole team can deploy. Your services then ship onto that infrastructure straight from Git. It deploys to your cloud accounts (AWS, GCP, Azure, Kubernetes) while providing a single control plane for the full software delivery lifecycle.

This documentation covers every shipped feature of the platform, from connecting your cloud accounts to deploying production workloads.

<!-- SCREENSHOT: Planton console dashboard
  Page: /dashboard
  Action: Show the main dashboard with resource cards, recent pipelines, and infra components
  Focus: Full dashboard view with summary cards and activity lists
  Alt: Planton console dashboard showing resource counts, recent pipelines, and infra component summary
-->

## Platform

Foundational concepts, resource hierarchy, and how the platform is organized.

[Get started with the platform](/docs/platform)

## Connections

Credential and integration management — connect cloud providers, Git providers, container registries, state backends, and Kubernetes clusters.

[Learn about Connections](/docs/connections)

## Infrastructure

Declarative infrastructure provisioning across cloud providers. Deploy individual Infra Components, compose them into Infra Charts, orchestrate with Infra Pipelines, and track execution through Infra Jobs.

[Explore Infrastructure](/docs/infrastructure)

## CI/CD

Application CI/CD from Git push to production deployment. Build with Buildpacks or Dockerfiles, deploy to Kubernetes, ECS, Cloud Run, or Cloudflare Workers, and manage secrets and ingress configuration.

[Explore CI/CD](/docs/ci-cd)

## Operations

Runtime operations gateway for Kubernetes pod management, log streaming, shell access, and multi-infra component browsing.

[Learn about Operations](/docs/operations)

## Runner

A single native binary deployed in your infrastructure that handles secure execution of IaC operations, reverse tunnel connectivity, and CloudOps request handling. Credentials never leave your environment.

[Learn about Runner](/docs/runner)

## Security

Credential isolation, encryption, authorization, audit trails, and the platform's security architecture.

[Learn about Security](/docs/security)

## Teams and Access

Team management, role-based permissions, and billing.

[Manage teams and access](/docs/teams-and-access)

## Quick Links

- [Getting Started](/docs/getting-started) — From signup to first deployment
- [CLI Reference](/docs/cli) — Install and use the Planton CLI
