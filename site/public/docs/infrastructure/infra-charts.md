---
title: "Infra Charts"
sidebar_title: "Charts"
description: "Reusable infrastructure templates that bundle multiple Infra Components into parameterized, dependency-aware packages you deploy with one command."
icon: infrastructure
order: 35
tags:
  - Infrastructure
  - Infra Charts
  - Templates
---

# Infra Charts

Deploying a production-ready environment on AWS typically involves ten or more interdependent resources: a VPC with subnets across multiple availability zones, NAT gateways, security groups, an Application Load Balancer with SSL termination, Route 53 DNS records, an ACM certificate, an ECS cluster with Fargate capacity, an ECR repository, and IAM roles. Getting the dependencies wrong — deploying the ALB before the VPC exists — produces cryptic errors or partial failures. Repeating this process for every new environment is time-consuming, error-prone, and inconsistent across teams.

Infra Charts solve this by packaging related Infra Components into a single, parameterized template. You provide a handful of values — domain name, region, availability zones — and the chart handles everything else: resource definitions, dependency ordering, and configuration wiring. What would take hours of manual work becomes a single command.

An Infra Chart is a blueprint, not a deployment. Creating or selecting a chart does not provision infrastructure. When you are ready to deploy, you create an [Infra Stack](/docs/infrastructure/infra-stacks) — a configured instance of the chart with your specific values — which triggers an [Infra Pipeline](/docs/infrastructure/infra-pipelines) that orchestrates the deployment of all resources in the correct order.

## Why Infra Charts Exist

### Repetitive Infrastructure Patterns

Teams build the same infrastructure patterns repeatedly. A development environment needs a VPC, subnets, security groups, and a container cluster. A staging environment needs the same pattern with different sizing. A new project needs the same pattern again, copied and adapted from the last one. Without a reuse mechanism, every environment is built from scratch.

### Hidden Dependencies

Infrastructure resources depend on each other in ways that must be orchestrated correctly. A load balancer needs a VPC and subnets. An ECS service needs a cluster, security groups, and a load balancer. Deploy them in the wrong order and you get errors. Traditional tools leave this coordination to you — manually sequencing deployments, chaining workflows, or writing custom scripts.

### No Standardization

Organizations develop best practices — security groups should restrict HTTP in production, VPCs should span multiple availability zones, all resources should be tagged for cost allocation — but there is no standard way to capture and enforce these patterns. Every project implements them differently, or forgets them entirely.

Infra Charts address all three problems: capture a pattern once, declare dependencies explicitly, and deploy consistently every time.

## Chart Structure

Every Infra Chart is a directory with three components, following a structure similar to Helm charts:

```
aws/ecs-environment/
├── Chart.yaml        # Metadata (name, description, links)
├── values.yaml       # Parameters with defaults
└── templates/        # Infra Component definitions with placeholders
    ├── network.yaml
    ├── ecs-cluster.yaml
    └── iam-role.yaml
```

**Chart.yaml** identifies the chart and provides metadata for discovery — name, description, icon, and links to documentation. Charts are scoped to a level of ownership: platform-level charts are available globally, organization-level charts are available to all environments in that organization.

**values.yaml** declares the configurable inputs. Each parameter has a name, type, description, and optional default value. Parameters appear as form fields in the web console when creating an Infra Stack from the chart.

**templates/** contains one or more YAML files defining catalog object manifests. Templates use Jinjava syntax for variable substitution and conditionals, and declare dependencies between resources so the platform knows what order to deploy them in.

## Templating with Jinjava

Infra Charts use Jinjava (a Java implementation of Jinja2) for dynamic configuration. Templates support variable substitution, conditionals, loops, and filters.

**Variable substitution** injects parameter values into resource definitions:

```yaml
metadata:
  name: "{{ values.env }}-vpc"
spec:
  region: "{{ values.aws_region }}"
```

**Conditionals** include or exclude entire resources based on parameter values:

```yaml
{% if values.httpsEnabled | bool %}
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCertManagerCert
metadata:
  name: "{{ values.env }}-alb-cert"
spec: {}
{% endif %}
```

This means a single chart can handle multiple deployment profiles. An ECS environment chart with DNS and HTTPS toggles can produce a minimal development setup or a fully secured production environment from the same template.

## Cross-Resource Dependencies

The key feature of Infra Charts is explicit dependency declaration. Resources within a chart can reference outputs of other resources — for example, a security group that needs the VPC's ID, or an ALB that needs subnet IDs from the VPC:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAlb
metadata:
  name: "{{ values.env }}-alb"
spec:
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: "{{ values.env }}-vpc"
        fieldPath: status.outputs.publicSubnets.[0].id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: "{{ values.env }}-sg"
        fieldPath: status.outputs.securityGroupId
```

These references create a directed acyclic graph (DAG). The platform deploys resources in topological order — VPC first, then security groups and subnets in parallel, then the ALB that depends on both. You never need to manually coordinate deployment ordering.

## Example: An ECS Environment Chart

An organization's chart for an AWS ECS environment shows how charts work in practice. It might compose a complete container hosting environment:

| Resource | Purpose |
|----------|---------|
| VPC | Network isolation with public/private subnets across 2 availability zones |
| Security Group | HTTP/HTTPS ingress rules, all egress |
| Route 53 Zone | DNS hosting (conditional on parameter) |
| ACM Certificate | SSL/TLS certificate (conditional on HTTPS) |
| Application Load Balancer | Traffic distribution with optional SSL termination |
| ECS Cluster | Container orchestration with Fargate |
| ECR Repository | Container image storage |
| IAM Role | Task execution permissions |
| ECS Service | Running container with ALB integration |

Such a chart takes a handful of parameters (availability zones, domain name, service name, service port, toggles for DNS and HTTPS) and produces every resource above in dependency order.

<!-- SCREENSHOT: Infra Chart detail page
  Page: /resource/infra-hub/infra-chart/[infraChartId]
  Action: Show a chart with template files and parameters visible
  Focus: The template YAML viewer and parameter list
  Alt: Infra Chart detail page showing template YAML files and a list of configurable parameters
-->

## Your Organization's Charts

Every chart belongs to an organization. Planton ships none of its own: a chart is your architecture, written for your clouds and your environments. The fastest way to your first one is the prompt on the Charts tab: paste it into Claude Code, Codex or Cursor, and your agent asks what the chart should stand up, builds it until `planton chart build` is clean, and publishes it with `planton chart publish`.

You can also write a chart by hand: follow the directory structure above (Chart.yaml, values.yaml, templates/), encode your organization's standards into the templates, and publish it to your organization.

Common customizations include enforcing tagging policies, requiring HTTPS in production environments, setting minimum replica counts, and restricting VPC CIDR ranges to organizational standards.

## Chart Versions

Planton counts a chart's versions; you never write one. A chart is **v1** when first published and becomes the next version each time its templates or `values.yaml` change, so a version always means the same files — each carries a digest Planton computes from them. Publishing unchanged files keeps the version, and `planton chart publish` prints the version it made. Chart.yaml has no `version:` line.

A stack installed from a catalog chart records the chart and the version it took ("app-infra v3"). When the chart publishes a newer version, that stack reads **Update Available**, and **Upgrade** moves it there, keeping its values and any file it customised: a file both you and the chart changed asks **Keep Mine** or **Take v5**. Publishing never changes a deployed stack on its own. See [Infra Stacks](/docs/infrastructure/infra-stacks#upgrade).

<!-- SCREENSHOT: Infra Chart list
  Page: /resource/infra-hub/infra-chart
  Action: Show the organization's charts
  Focus: The chart list with names, descriptions, and versions
  Alt: Infra Chart list showing an organization's charts with their descriptions and versions
-->

## Charts vs. Direct Infra Components

Planton supports both deploying individual Infra Components and deploying bundled charts. Choose based on the situation:

| Scenario | Recommended Approach |
|----------|---------------------|
| Complete environment (dev, staging, production) | Infra Chart — coordinated deployment with dependency handling |
| Single resource (one S3 bucket, one DNS record) | Direct Infra Component — no orchestration overhead |
| Standardized patterns across teams | Infra Chart — encode best practices in templates |
| Quick experiments | Direct Infra Component — fast iteration, easy to delete |
| Complex multi-resource dependencies | Infra Chart — automatic DAG-based ordering |

You can start with a chart for initial deployment, then manage individual Infra Components afterward for incremental changes.

## Using the CLI

```bash
# Ground yourself before writing templates: the full schema reference for any
# resource kind or platform API, fully offline (kubectl-explain style)
planton explain aws-vpc
planton explain infra-chart

# Drill into one field with a dotted path of the exact YAML keys
planton explain aws-vpc.spec.instanceTenancy

# List every kind the CLI can explain (platform APIs + catalog kinds)
planton explain --list

# Build a chart from a local directory (renders and validates, reports issues)
planton chart build ./my-chart

# Build and print the rendered documents (the params applied) to stdout
planton chart build ./my-chart --show

# Machine-readable report for CI pipelines and agents (exit code 0 = valid,
# 1 = chart has build errors, 2 = the check could not run)
planton chart build ./my-chart --output-format json

# Build a variant without editing the chart: override params for this one
# build (repeatable; the value is parsed as YAML, like values.yaml)
planton chart build ./my-chart --set dns_enabled=true --set region=eu-west-1

# Compile an environment's values file before installing it
planton chart build ./my-chart -f values-prod.yaml --show

# Publish a chart to the organization's chart registry; prints the version it made
planton chart publish ./my-chart

# List all charts available for the organization, each with its version
planton chart list

# Install from the catalog: the stack records the chart and its version
planton chart install my-stack app-infra --env prod -f values-prod.yaml -m "why"

# Plan what installing would do, and apply nothing
planton chart install my-stack ./my-chart --env prod -f values-prod.yaml --dry-run

# Install: plan, review, and apply exactly that plan
planton chart install my-stack ./my-chart --env prod -f values-prod.yaml -m "why"
```

The `chart install` command is the primary entry point for deploying a chart. Like `planton infra stack apply`, it plans first — what installing creates, or changes in the stack of that name — shows the plan, asks, and applies exactly that plan as the stack's next revision, starting an Infra Pipeline. The stack keeps the chart's `values.yaml` and only the values you set. The chart is a local folder or one of your organization's charts by name. From a published chart, Planton copies the chart's files itself and records the chart and its version; installing onto a stack that already follows that chart is an upgrade, and a file both changed is settled with `planton infra stack upgrade`.

### How parameter values resolve

A parameter's value in the chart's `values.yaml` is its default. An install or build changes parameters by name, in this order: the chart's `values.yaml`, then each `-f` file in the order given, then `--set-file name=path`, then `--set name=value`. Later sources win.

A values file has the `values.yaml` shape and lists only the parameters it changes; every parameter it does not name keeps the chart's value:

```yaml
# values-prod.yaml
params:
  - name: env_name
    value: prod
  - name: replicas
    value: 3
```

A name the chart does not declare is refused with the list of the chart's parameters, whether it comes from a file, `--set` or `--set-file`. Every command prints where each parameter's value came from. An installed stack stores only the values you set; every other parameter takes the default from the `values.yaml` the stack carries. `chart build` shows the render; `chart install --dry-run` (or `planton infra stack plan`) shows the plan — each resource as Create, Update, Destroy, Keep and the rest — and applies nothing. With `--output-format json`, a plan prints its machine record, the rendered build included.

## Related Documentation

- [Infra Components](/docs/infrastructure/infra-components) — The individual resources that charts compose
- [Catalog Kinds](/docs/infrastructure/catalog-kinds) — The catalog of resource types available in charts
- [Infra Stacks](/docs/infrastructure/infra-stacks) — Running instances of Infra Charts
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — How chart deployments are orchestrated
- [Open Source](/docs/infrastructure/open-source) — The open-source foundation that provides resource APIs and cross-resource dependencies
