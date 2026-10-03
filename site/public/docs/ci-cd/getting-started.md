---
title: "Getting Started with Service Hub"
sidebar_title: "Getting Started"
description: "Deploy your first service — from connecting a Git provider to watching the pipeline build and deploy your code."
icon: getting-started
order: 15
tags:
  - Getting Started
  - Service Hub
---

# Getting Started with Service Hub

This guide walks you through deploying your first service with Service Hub. By the end, you will have connected a Git repository, created a Service, and watched a pipeline build and deploy your code.

## Prerequisites

- A Planton account with an organization
- An environment created in your organization (see [Platform Getting Started](/docs/platform/getting-started))
- A GitHub repository with application code
- A container registry connection (see [Container Registries](/docs/connections/container-registries))

## Step 1: Connect Your Git Provider

Service Hub needs access to your Git repositories to clone code and to learn about pushes. Navigate to **Connections** and add your GitHub connection.

A GitHub connection signs in one of two ways. As a **GitHub App** (hosted Planton), GitHub delivers push events to Planton through the App's own subscription and pipelines start automatically — nothing is registered on your repository. As the **sign-in on this machine** (a laptop or a self-hosted server), the connection stores no token and Planton watches the repository for new commits to start runs itself.

See [Git Providers](/docs/connections/git-providers) for detailed setup instructions.

<!-- SCREENSHOT: Service creation wizard
  Page: /orgs/{org}/services (create flow)
  Action: Show the service creation form with repository selection and build method configuration
  Focus: The form fields for service name, repository, build method, and deployment target
  Alt: Service creation wizard showing repository selection, build method choice, and deployment target configuration
-->

## Step 2: Create a Service

Navigate to **Service Hub** in the web console. If this is your first service, the deploy wizard starts automatically. Otherwise, click **Add New Service**.

The wizard walks you through the configuration:

1. **Select your Git provider** — Choose GitHub (GitLab support is coming soon).
2. **Pick a repository** — Browse your connected repositories and select the one to deploy.
3. **Name your service** — Confirm the repository and give the service a name.
4. **Choose a package type** — Container image (for Kubernetes, ECS, Cloud Run) or Cloudflare Worker script.
5. **Configure the build** — Select Buildpacks (auto-detect language) or Dockerfile. Configure the container registry and optionally set project root, trigger paths, and pipeline branches.
6. **Configure deployment** — Choose Git-based (Kustomize overlays in your repo) or UI-based (configure deployment targets directly in the wizard). Add environments with their cloud provider and resource type.
7. **Review and create** — Confirm the configuration and create the service.

After creation, Planton configures a webhook on your repository and triggers the first pipeline.

See [What is a Service?](/docs/ci-cd/what-is-a-service) for a full explanation of each configuration area.

## Step 3: Monitor the Pipeline

The first pipeline starts immediately. Navigate to the **Pipelines** tab on your service detail page to watch progress.

The pipeline progresses through two stages — **Build** (clone, build artifact, push to registry) and **Deploy** (provision infra components for each environment). Build logs stream in real time. Each deployment environment creates an Infra Job that you can inspect for detailed provisioning output.

See [Pipelines](/docs/ci-cd/pipelines) for the full pipeline model, trigger types, and manual approval gates.

## Step 4: Push Code

From this point forward, every push to a configured branch triggers a new pipeline automatically. Make a change, push it, and watch the build-and-deploy cycle run without manual intervention.

## What to Explore Next

- **[Deployment Targets](/docs/ci-cd/deployment-targets)** — Configure where your service runs (Kubernetes, ECS, Cloud Run, Workers)
- **[Build Methods](/docs/ci-cd/build-methods)** — Understand Buildpacks vs Dockerfile builds
- **[Deployment Environments](/docs/ci-cd/deployment-environments)** — Control which environments a service deploys to
- **[Secrets](/docs/secrets)** — Inject secrets and variables into deployments
- **[Ingress](/docs/ci-cd/ingress)** — Make your service publicly accessible with a DNS domain
- **[Monorepo Support](/docs/ci-cd/monorepo-support)** — Deploy multiple services from a single repository
