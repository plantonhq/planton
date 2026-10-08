---
title: "Getting Started with Infrastructure"
sidebar_title: "Getting Started"
description: "Deploy your first Infra Component through Infrastructure — from connecting a cloud provider to monitoring the Infra Job."
icon: getting-started
order: 15
tags:
  - Getting Started
  - Infrastructure
---

# Getting Started with Infrastructure

This guide walks you through deploying your first piece of infrastructure with Infrastructure. By the end, you will have connected a cloud provider, created an Infra Component, and watched an Infra Job provision it.

## Prerequisites

- A Planton account with an organization
- An environment created in your organization (see [Platform Getting Started](/docs/platform/getting-started))
- Access to a cloud provider account (AWS, GCP, or Azure) or a Kubernetes cluster

## Step 1: Connect a Cloud Provider

Before Infrastructure can provision infrastructure, it needs credentials for your cloud provider. Navigate to **Connect** in the web console and add your provider credentials.

Each provider has its own connection type. AWS, GCP, and Azure connections can be keyless — your account trusts Planton's identity issuer, and no credential is stored (see [Keyless Cloud Connections](/docs/connections/keyless-cloud-connections)) — or use access keys, service account keys, or service principals. Kubernetes connections use kubeconfig files.

See the [Connections](/docs/connections) section for detailed setup instructions for each provider.

<!-- SCREENSHOT: Connect page with provider options
  Page: /orgs/{org}/connections
  Action: Show the connections page with provider types visible
  Focus: The provider type selection or existing connection list
  Alt: Connect page showing available cloud provider connection types
-->

## Step 2: Map Credentials to Your Environment

After adding a connection, authorize it for the environment where you want to deploy. This tells Planton which credentials to use when provisioning resources in that environment.

If this is your first connection for a provider, you can also set it as the default — so you do not need to specify the connection explicitly every time you create a resource.

See [Environment Mappings](/docs/connections/environment-mappings) and [Default Connections](/docs/connections/default-connections) for details.

## Step 3: Browse the Catalog

Navigate to the **Infra Catalog** in the web console to see what you can deploy. The catalog shows all available resource types — AWS VPCs, GCP Cloud SQL instances, Kubernetes deployments, and more — organized by provider.

Select a resource type to start configuring it for deployment.

See [Catalog Kinds](/docs/infrastructure/catalog-kinds) for an overview of the full catalog.

<!-- SCREENSHOT: Infra Catalog
  Page: /platform/deployment-store
  Action: Show the catalog with resource types from multiple providers
  Focus: The kind grid with provider branding and deploy buttons
  Alt: Infra Catalog showing available infrastructure types organized by cloud provider
-->

## Step 4: Create an Infra Component

Select a resource type from the catalog, configure it for your target environment, and deploy. The web console guides you through the configuration — selecting the environment, setting resource-specific options, and reviewing before submission.

You can also create an Infra Component from the CLI using a YAML manifest:

```bash
planton create -f my-resource.yaml
```

See [Infra Components](/docs/infrastructure/infra-components) for the full lifecycle and configuration options.

## Step 5: Monitor the Infra Job

When you submit the Infra Component, an [Infra Job](/docs/infrastructure/infra-jobs) is created automatically. The job initializes the IaC module, refreshes state, previews changes, and applies them. Watch progress in real time from the web console or CLI:

```bash
planton infra job stream-progress-events <infra-job-id>
```

The Infra Job detail page shows each operation step with its status — initialize, refresh, preview, apply — along with resource-level logs showing exactly what is being created.

<!-- SCREENSHOT: Infra Job progress
  Page: /resource/infra-hub/infra-job/[infraJobId]
  Action: Show an Infra Job in progress with operation steps visible
  Focus: The operation steps with status indicators
  Alt: Infra Job progress page showing initialize completed, refresh completed, preview in progress
-->

## What to Explore Next

- **[Infra Charts](/docs/infrastructure/infra-charts)** — Bundle multiple resources into reusable templates
- **[Infra Stacks](/docs/infrastructure/infra-stacks)** — Deploy coordinated multi-resource environments
- **[Infra Pipelines](/docs/infrastructure/infra-pipelines)** — Orchestrate deployments with dependency ordering
- **[Flow Control](/docs/infrastructure/flow-control)** — Add governance policies for production deployments
