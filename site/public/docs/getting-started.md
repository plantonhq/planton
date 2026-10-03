---
title: "Getting Started"
description: "Choose your starting point — deploy infrastructure, ship applications, or understand how the platform is organized."
icon: getting-started
order: 1
tags:
  - Getting Started
  - Overview
---

# Getting Started

New to Planton? Start with the [Platform Getting Started Guide](/docs/platform/getting-started) to create your account, set up an organization, connect a cloud provider, and deploy your first resource.

Already set up? Choose a path based on what you want to do.

## Deploy Infrastructure

1. [Connect your cloud provider](/docs/connections/cloud-providers) — link your AWS, GCP, or Azure account
2. [Browse the deployment catalog](/docs/infrastructure/catalog-kinds) — see what you can deploy
3. [Deploy an Infra Component](/docs/infrastructure/infra-components) — provision a VPC, database, Kubernetes cluster, or any supported resource
4. [Compose with Infra Charts](/docs/infrastructure/infra-charts) — deploy a coordinated set of resources as a single unit

## Deploy Applications

1. [Connect your Git provider](/docs/connections/git-providers) — link your GitHub or GitLab organization
2. [Create a Service](/docs/ci-cd/what-is-a-service) — connect a repository and configure build settings
3. [Push code and deploy](/docs/ci-cd/pipelines) — trigger automated builds and deployments
4. [Configure ingress](/docs/ci-cd/ingress) — set up domains and routing for your service

## Work From Your Coding Agent

1. [Install the Planton skills](/docs/coding-agents) — one command for Cursor, Claude Code, Codex, and any Agent Skills host
2. [Install the CLI](/docs/cli) — validation and schema lookups offline, deploys after one sign-in
3. Ask your agent for what you need — it composes under `infrastructure/`, validates, and asks before applying

## Understand the Platform

- [Platform Overview](/docs/platform) — what Planton is and how it is organized
- [Platform Tour](/docs/platform/platform-tour) — walkthrough of the console interface
- [Core Concepts](/docs/platform/core-concepts) — key terminology and how concepts connect
- [Resource Hierarchy](/docs/platform/resource-hierarchy) — organizations, environments, and resources

## Documentation Sections

| Section | What It Covers |
|---------|---------------|
| [Platform](/docs/platform) | Resource hierarchy, core concepts, and console navigation |
| [Connections](/docs/connections) | Credentials and integrations for cloud providers, Git, registries, and state backends |
| [Infrastructure](/docs/infrastructure) | Infrastructure provisioning — Infra Components, Infra Charts, Infra Pipelines, Infra Jobs |
| [CI/CD](/docs/ci-cd) | Application CI/CD — Services, Pipelines, build methods, deployment targets |
| [Secrets](/docs/secrets) | Secrets management, variables, and secret backends |
| [Operations](/docs/operations) | Runtime operations — Kubernetes pod management, log streaming, resource browsing |
| [Runner](/docs/runner) | Secure execution agent deployed in your infrastructure |
| [Teams and Access](/docs/teams-and-access) | Team management, role-based permissions, and billing |
| [CLI](/docs/cli) | Install and use the Planton command-line interface |
| [Coding Agents](/docs/coding-agents) | Install the Planton skills into Cursor, Claude Code, and other coding agents |
