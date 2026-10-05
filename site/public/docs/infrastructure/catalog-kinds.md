---
title: "Catalog Kinds"
description: "The catalog of infrastructure types available in Planton — from AWS VPCs to Kubernetes workloads — across all supported cloud providers."
icon: concepts
order: 25
tags:
  - Infrastructure
  - Catalog Kinds
  - Catalog Kinds
---

# Catalog Kinds

Planton supports over 700 types of infrastructure across 8 cloud providers. Each type — an AWS VPC, a GCP Cloud SQL instance, a Kubernetes deployment, a Cloudflare DNS zone — is a catalog kind. When you browse the Infra Catalog in the web console, you are browsing catalog kinds with their descriptions, icons, and provider information.

## Why a Unified Catalog

Each cloud provider has its own console, API, and documentation for provisioning resources. An engineer managing infrastructure across AWS, GCP, and Kubernetes would normally switch between three different interfaces with three different sets of conventions.

The Infra Catalog brings all of these into a single browsable catalog. You see everything you can deploy in one place, organized by provider, with consistent configuration and lifecycle management regardless of the underlying cloud. When you select a catalog kind, Planton handles the rest — resolving the correct IaC module, connecting to the right provider credentials, and tracking the resource through its lifecycle.

## What You Can Deploy

### AWS

Amazon Web Services resources including VPCs, RDS databases (instances and clusters), EKS clusters, ECS services, Lambda functions, S3 buckets, Application Load Balancers, Route 53 DNS zones, and more.

### Google Cloud Platform

GCP resources including GKE clusters, Cloud SQL databases, Cloud Run services, VPC networks, Cloud Functions, GCS buckets, Cloud DNS zones, and more.

### Microsoft Azure

Azure resources including AKS clusters, SQL databases, Storage Accounts, and more.

### Kubernetes

Kubernetes-native workloads and add-ons that deploy on top of existing clusters. This includes managed databases (PostgreSQL, Redis, MongoDB), message brokers (Kafka, NATS), identity platforms (Keycloak), service meshes (Istio), certificate management (cert-manager), monitoring (Prometheus), and many more.

Unlike cloud provider resources that provision infrastructure directly, Kubernetes resources deploy applications and services onto clusters you have already connected.

### Other Providers

Additional providers include Cloudflare, DigitalOcean, Auth0, and OpenFGA.

<!-- SCREENSHOT: Infra Catalog
  Page: /infra-catalog
  Action: Show the catalog filtered by a specific provider (e.g., AWS)
  Focus: The grid of available kinds with icons and descriptions
  Alt: Infra Catalog showing AWS resources including VPC, RDS, EKS, and Lambda
-->

## Browsing the Catalog

### Web Console

Navigate to the **Infra Catalog** in the web console to browse the full catalog. Resources are displayed as cards with icons, provider branding, and descriptions. You can filter by provider to narrow the view.

When you select a catalog kind, Planton takes you to the creation form where you configure the Infra Component for your target environment and deploy it.

### CLI

```bash
# List every catalog kind
planton explain --list

# Get details about a specific catalog entry
planton get catalog-kind <catalog-kind-id>

# List all catalog entries
planton list catalog-kind
```

## From Catalog to Infra Component

The catalog is the starting point. The deployment flow is:

1. **Browse** the catalog and select a resource type.
2. **Configure** the resource for your target environment — provider credentials, region, resource-specific settings.
3. **Deploy** — Planton creates an [Infra Component](/docs/infrastructure/infra-components) and runs an [Infra Job](/docs/infrastructure/infra-jobs) to provision the infrastructure.

You can also use resource types in [Infra Charts](/docs/infrastructure/infra-charts) to compose multiple resources into reusable templates, and deploy them together through [Infra Stacks](/docs/infrastructure/infra-stacks).

## Service-Deployable Kinds

Some resource types can serve as deployment targets for [CI/CD](/docs/ci-cd). These include Kubernetes Deployments, Kubernetes StatefulSets, AWS ECS Services, GCP Cloud Run, and Cloudflare Workers. When you configure a service's deployment targets, you are selecting from these service-deployable resource types. See [Deployment Targets](/docs/ci-cd/deployment-targets) for details.

## Extending the Catalog

New resource types are added through [Planton open source](/docs/infrastructure/open-source) — the open-source multi-cloud foundation that Planton builds on. Each resource type requires an API definition and an IaC module (Pulumi, Terraform, or OpenTofu). Once defined in the open-source core, the resource type becomes available in Planton's catalog automatically.

## Related Documentation

- [Infra Components](/docs/infrastructure/infra-components) — Deployed instances of resource types from the catalog
- [Open Source](/docs/infrastructure/open-source) — The open-source foundation that defines resource types
- [Infra Charts](/docs/infrastructure/infra-charts) — Composing multiple resource types into reusable templates
