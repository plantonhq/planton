---
title: "Platform Tour"
description: "A walkthrough of the Planton console — sidebar navigation, context selector, Catalog Kind Store, and what each section contains."
icon: tour
order: 20
tags:
  - Platform Tour
  - Console
  - Navigation
---

# Platform Tour

This page walks through the Planton web console section by section, explaining what each area does and how to navigate between them. Use it as a reference during your first week on the platform.

## Console Layout

The console has three main areas:

- **Header** — context selector (top-left), search, Catalog Kind Store and IaC Module Registry (top-right)
- **Sidebar** — primary navigation to all platform sections
- **Main content** — the active page, which changes based on your sidebar selection and context

<!-- SCREENSHOT: Console layout overview
  Page: /dashboard
  Action: Show the default console view after login with sidebar expanded
  Focus: Full page showing header, left sidebar, and main content area
  Alt: Planton console layout with header navigation, left sidebar menu, and main content dashboard
-->

## Context Selector

The context selector sits in the top-left of the header, next to the Planton logo. It shows your current position in the resource hierarchy:

```
Acme Corp / production
    ↑            ↑
Organization   Environment
```

Clicking the context selector opens a dropdown that lists your organization and its environments. Selecting a different environment changes what resources, services, and connections are visible throughout the console. Organization-level pages (Connections, Billing, Settings) remain the same regardless of environment selection.

<!-- SCREENSHOT: Context selector expanded
  Page: /dashboard (header area)
  Action: Click context selector to show expanded dropdown with org and environment tree
  Focus: The expanded context selector dropdown
  Alt: Context selector dropdown showing organization name with nested environment list
-->

The CLI equivalent is `planton context set`:

```bash
# Set your working context
planton context set --org acme-corp --env production

# View current context
planton context get
```

## Sidebar Sections

The sidebar contains seven sections. Here is what each one does.

### Dashboard

The landing page after login. For new users, it shows the getting-started checklist (8 tasks across 4 groups — Foundation, Infrastructure, Applications, Team and Operations). For active organizations, it shows resource summaries, recent deployments, and notifications.

<!-- SCREENSHOT: Dashboard
  Page: /dashboard
  Action: Show the dashboard for an active organization
  Focus: The main content area with summary cards
  Alt: Planton dashboard showing resource summaries and recent deployment activity
-->

### Infra Hub

The infrastructure management section. The Infra Hub route activates for Infra Components, Infra Stacks, and environments.

**Infra Components tab** — lists all deployed infrastructure in the current environment. Each row shows the resource name, kind (e.g., AWS VPC, GCP GKE Cluster), status, and last deployment timestamp. Click a resource to see its full configuration, Infra Job history, and outputs.

**Infra Stacks tab** — lists Infra Chart deployments. Each Infra Stack groups multiple related Infra Components deployed as a coordinated unit. The detail view includes a DAG visualization showing resource dependencies and deployment progress.

**Environments tab** — lists environments in the organization with creation details.

<!-- SCREENSHOT: Infra Hub Infra Components
  Page: /orgs/{org}/infra-components
  Action: Show the Infra Components tab with at least 2-3 resources listed
  Focus: The resource list table with status indicators
  Alt: Infra Hub showing Infra Components tab with deployed infrastructure listed in a table
-->

[Learn more about Infrastructure](/docs/infrastructure)

### Service Hub

The application deployment section. Service Hub manages the full lifecycle from Git repository to running service.

The Service Hub route covers several tabs:

- **Services** — list of connected Git repositories configured for deployment
- **Variables** — environment variables grouped and scoped to services
- **Secrets** — encrypted secrets managed through the Secrets Manager
- **Domains** — DNS domains and ingress configuration
- **Package Repositories** — NPM and Maven registries for build-time resolution
- **Hosting Providers** — deployment targets for Cloudflare Workers (Wrangler credentials)
- **Environment Promotion** — promotion policies for moving deployments between environments

<!-- SCREENSHOT: Service Hub
  Page: /orgs/{org}/services
  Action: Show the Services tab with at least one service listed
  Focus: The services list with build and deployment status
  Alt: Service Hub showing services list with Git repository connections and deployment status
-->

[Learn more about CI/CD](/docs/ci-cd)

### Agent Fleet

An AI agent management area for defining agents, skills, and reviewing execution sessions. Agent Fleet operates independently from infrastructure and application deployment.

### Connections

Where you manage credentials and integrations with external services. The Connections page organizes providers into categories: Infrastructure (AWS, GCP, Azure, and others), DevOps Pipeline (GitHub, GitLab, container registries), Infrastructure as Code (state backends), and Managed Services.

The Connections route also includes a **Runners** tab, where you manage Runner deployments — the secure execution agents that run in your infrastructure.

<!-- SCREENSHOT: Connections page
  Page: /orgs/{org}/connections
  Action: Show the Connections page with provider cards
  Focus: The grid of provider cards organized by category
  Alt: Connections page showing provider cards for AWS, GCP, Azure, GitHub, and other integrations
-->

[Learn more about Connections](/docs/connections) | [Learn more about Runner](/docs/runner)

### Billing

Subscription and usage management. Contains two sub-pages:

- **Plans** — view available subscription tiers (Free, Plus, Pro) and your current plan
- **Subscription** — manage payment methods and view billing history

[Learn more about Billing](/docs/teams-and-access/billing)

### Settings

Organization-wide configuration with three tabs:

- **General** — organization name, description, contact email, and logo
- **Manage Members** — invite new members via email, view current members and their roles, manage pending invitations
- **Teams** — create teams, add members to teams, and manage team-level permissions

<!-- SCREENSHOT: Settings page
  Page: /orgs/{org}/settings
  Action: Show the Settings page with the General tab active
  Focus: The tab bar (General, Manage Members, Teams) and form fields
  Alt: Organization settings page showing General tab with organization name and description fields
-->

[Learn more about Teams and Access](/docs/teams-and-access)

## Header Actions

Beyond the context selector, the header provides quick access to two important features:

### Catalog Kind Store

Click the store icon in the header (right side) to open the catalog of deployable infrastructure kinds. The store has two sections:

- **Catalog Kinds** — individual Infra Component templates (e.g., AWS VPC, GCP GKE Cluster, Azure AKS). Filter by cloud provider and search by name. Click **Deploy** on any kind to start the deployment wizard.
- **Infra Charts** — pre-composed collections of Catalog Kinds that deploy together as a coordinated unit (e.g., an AWS ECS environment with VPC, cluster, load balancer, and DNS).

<!-- SCREENSHOT: Catalog Kind Store
  Page: /platform/deployment-store
  Action: Show the kind catalog with provider filter
  Focus: The kind grid with deploy buttons
  Alt: Catalog Kind Store showing infrastructure kinds filterable by cloud provider
-->

### IaC Module Registry

Click the IaC Module Registry icon in the header to browse the Pulumi, Terraform, and OpenTofu modules that back each Catalog Kind. This is useful for understanding what infrastructure-as-code runs behind a deployment, or for referencing module parameters.

## Navigation Patterns

### Context-Driven Content

The console adapts based on your current context:

- **No organization** — dashboard prompts you to create one
- **Organization selected, no environment** — organization-level pages (Connections, Billing, Settings) are accessible; environment-scoped pages prompt you to create or select an environment
- **Organization and environment selected** — all pages show content scoped to that environment

### Finding Resources

There are three ways to find what you need:

1. **Sidebar navigation** — click the section name to see all resources of that type
2. **Catalog Kind Store** — browse or search the catalog when you want to deploy something new
3. **Context selector** — switch environments to see resources in a different deployment stage

## Related Documentation

- [Getting Started](/docs/platform/getting-started) — step-by-step guide from signup to first deployment
- [Resource Hierarchy](/docs/platform/resource-hierarchy) — how organizations, environments, and resources are structured
- [Core Concepts](/docs/platform/core-concepts) — reference for key platform terminology
