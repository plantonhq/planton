---
title: "Infra Projects"
sidebar_title: "Projects"
description: "Configured instances of Infra Charts or Git repositories that deploy and manage collections of Cloud Resources through automated pipelines."
icon: infrastructure
order: 40
tags:
  - Infrastructure
  - Infra Projects
  - Deployment
---

# Infra Projects

An Infra Chart defines what you *can* deploy — a reusable template with parameterized placeholders. But a template alone does not answer the questions that matter for a specific deployment: which region, which domain name, how many availability zones, which environment. An Infra Project captures those answers. It is a configured instance of a template with your specific parameter values filled in, ready to deploy as real infrastructure.

The relationship mirrors how Helm works: an Infra Chart is to an Infra Project as a Helm chart is to a Helm release. The chart is reusable; the project is the deployed instance with specific values. This separation enables teams to capture infrastructure patterns once and instantiate them many times — across different environments, regions, or customers — each instance tracked as a distinct, versioned project.

## Why Infra Projects Exist

Before Infra Projects, the platform had Infra Charts for reusability but no persistent record of each deployment instance. Consider an organization that deploys five ECS environments from the same chart — development, staging, production-us, production-eu, and production-asia. Each uses different parameter values. Without a project record, fundamental questions have no answers:

- Which parameters deployed production-us last month?
- Did staging update to the latest chart version?
- Who changed the load balancer domain in production-eu?
- Can we redeploy production-asia with last week's configuration?

Infra Projects solve this by introducing a persistent, versioned record of every template instantiation. Each project captures the source template, the exact parameter values used, the rendered Cloud Resource manifests, the dependency graph, and a link to every pipeline that deployed it. Pipelines are transient — they run and complete. Projects persist as the source of truth for what was configured and when.

## Two Source Models

Infra Projects support two distinct patterns for sourcing infrastructure configurations.

### Chart-Based

The most common approach. You select an Infra Chart from the catalog, provide parameter values for your target environment, and the project renders the chart templates into concrete Cloud Resource manifests.

When the project is created, the complete template is copied from the chart into the project. This ensures that chart updates do not break or silently change existing projects — each project is a self-contained snapshot of the template it was created from.

The web console provides a creation wizard: browse the chart catalog, fill in the parameter form (auto-generated from the chart's parameter definitions), preview the rendered output, and click "Create and Deploy."

<!-- SCREENSHOT: Infra Project creation from chart
  Page: /[org]/infra-projects/create/[infraChartId]
  Action: Show the creation wizard with chart parameters form
  Focus: The parameter form and environment selector
  Alt: Infra Project creation wizard showing chart parameters and environment selection for an AWS ECS environment chart
-->

### Git-Based

For teams that manage infrastructure definitions in version control. The project points to a Git repository containing Cloud Resource YAML manifests. Changes pushed to the repository trigger Infra Pipelines automatically via webhooks.

Git-based projects support the same pipeline controls as chart-based projects — you can disable automatic pipelines, disable deployments while keeping builds, or enable pull request deployments for preview environments.

## Automatic Pipeline Triggering

Creating or updating an Infra Project automatically triggers an [Infra Pipeline](/docs/infrastructure/infra-pipelines). The pipeline takes the project's Cloud Resource dependency graph and executes deployments in the correct order. There is no separate "apply" step — the platform assumes that creating a project implies intent to deploy.

For the rare cases where you want to create a project without deploying, pipeline controls let you disable automatic triggering:

- **Disable pipelines**: Stop all automatic pipeline triggering entirely
- **Disable deployments**: Allow pipelines to build and validate but skip the deployment stage
- **Enable pull request deployments**: Create preview deployments for pull requests (Git-based projects only)

## Dependency Graph Visualization

Each Infra Project maintains a dependency graph that visualizes the relationships between its Cloud Resources. The web console displays this as an interactive DAG where you can see:

- Resource nodes with their current deployment status (queued, running, succeeded, failed)
- Dependency edges showing which resources must be deployed before others
- Real-time updates as a pipeline progresses through the graph

This visualization makes it easy to understand the deployment topology and diagnose failures — if a resource fails, you can immediately see which downstream resources are blocked.

<!-- SCREENSHOT: Infra Project DAG visualization
  Page: /[org]/infra-project/[infraProjectSlug]
  Action: Show the project detail page with the DAG tab active
  Focus: The dependency graph visualization showing resources and their connections
  Alt: Infra Project DAG visualization showing an AWS VPC connected to subnets, security groups, and an RDS instance
-->

## Project Lifecycle

### Create

Create a project from a chart with the CLI:

```bash
planton chart install my-project ./my-chart -f values.yaml
```

Or from the web console's chart catalog by selecting a chart, filling in parameters, and clicking "Create and Deploy."

For Git-based projects, configure the repository source through the web console or CLI.

### Redeploy

Trigger a new pipeline without changing the project's configuration. Useful for drift correction (re-applying desired state after manual cloud console changes), retrying after a failed deployment, or re-running after fixing external issues like quota limits or permissions:

```bash
planton infra project deploy <project-name-or-id>
```

### Undeploy

Destroy all Cloud Resources owned by the project without deleting the project record. The project configuration is preserved and can be redeployed later — useful for temporarily tearing down infrastructure to save costs:

```bash
planton infra-project undeploy <project-name-or-id>
```

### Purge

Destroy all Cloud Resources and delete the project record permanently:

```bash
planton infra-project purge <project-name-or-id>
```

Deleting a project does not automatically destroy its infrastructure. You must explicitly undeploy first if the infrastructure should be removed. This is a deliberate safety measure — deleting a configuration record should never accidentally destroy production resources.

## Using the CLI

```bash
# Create a project from a chart (triggers deployment automatically)
planton chart install my-project ./chart-dir -f values.yaml

# Deploy an existing project (starts a deploy run and follows it)
planton infra project deploy <project-name-or-id>

# List pipelines for a project
planton infra-project infra-pipelines --project <project-id>

# Get the last pipeline for a project
planton infra-project last-pipeline --project <project-id>

# Undeploy (destroy resources, keep project)
planton infra-project undeploy <project-name-or-id>

# Purge (destroy resources and delete project)
planton infra-project purge <project-name-or-id>

# Get project details
planton get infra-project <project-id>
```

## When to Use Infra Projects vs. Direct Cloud Resources

| Scenario | Recommended Approach |
|----------|---------------------|
| Complete environment (dev, staging, production) | Infra Project — coordinated deployment, dependency handling, versioned configuration |
| Single resource needed quickly | Direct Cloud Resource — no orchestration overhead |
| Reproducible deployments | Infra Project — configuration captured as a versioned artifact |
| Quick experiments | Direct Cloud Resource — fast iteration, easy to delete |
| Production infrastructure | Infra Project — audit trails, parameter history, rollback capability |
| Adding to existing infrastructure | Direct Cloud Resource — targeted change without touching other resources |

## Related Documentation

- [Infra Charts](/docs/infrastructure/infra-charts) — The templates that chart-based projects instantiate
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — How project deployments are orchestrated
- [Cloud Resources](/docs/infrastructure/cloud-resources) — The resources that projects own and manage
- [Stack Jobs](/docs/infrastructure/stack-jobs) — The atomic execution units within pipelines
- [Flow Control](/docs/infrastructure/flow-control) — Governance policies that affect pipeline execution
