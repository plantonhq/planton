---
title: "Infra Charts"
sidebar_title: "Charts"
description: "Reusable infrastructure templates that bundle multiple Cloud Resources into parameterized, dependency-aware packages you deploy with one command."
icon: infrastructure
order: 35
tags:
  - Infrastructure
  - Infra Charts
  - Templates
---

# Infra Charts

Deploying a production-ready environment on AWS typically involves ten or more interdependent resources: a VPC with subnets across multiple availability zones, NAT gateways, security groups, an Application Load Balancer with SSL termination, Route 53 DNS records, an ACM certificate, an ECS cluster with Fargate capacity, an ECR repository, and IAM roles. Getting the dependencies wrong — deploying the ALB before the VPC exists — produces cryptic errors or partial failures. Repeating this process for every new environment is time-consuming, error-prone, and inconsistent across teams.

Infra Charts solve this by packaging related Cloud Resources into a single, parameterized template. You provide a handful of values — domain name, region, availability zones — and the chart handles everything else: resource definitions, dependency ordering, and configuration wiring. What would take hours of manual work becomes a single command.

An Infra Chart is a blueprint, not a deployment. Creating or selecting a chart does not provision infrastructure. When you are ready to deploy, you create an [Infra Project](/docs/infrastructure/infra-projects) — a configured instance of the chart with your specific values — which triggers an [Infra Pipeline](/docs/infrastructure/infra-pipelines) that orchestrates the deployment of all resources in the correct order.

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
└── templates/        # Cloud Resource definitions with placeholders
    ├── network.yaml
    ├── ecs-cluster.yaml
    └── iam-role.yaml
```

**Chart.yaml** identifies the chart and provides metadata for discovery — name, description, icon, and links to documentation. Charts are scoped to a level of ownership: platform-level charts are available globally, organization-level charts are available to all environments in that organization.

**values.yaml** declares the configurable inputs. Each parameter has a name, type, description, and optional default value. Parameters appear as form fields in the web console when creating an Infra Project from the chart.

**templates/** contains one or more YAML files defining Cloud Resource manifests. Templates use Jinjava syntax for variable substitution and conditionals, and declare dependencies between resources so the platform knows what order to deploy them in.

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

## Real-World Example: AWS ECS Environment

The AWS ECS Environment chart demonstrates how charts work in practice. It deploys a complete container hosting environment:

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

The chart takes about ten parameters (availability zones, domain name, service name, service port, feature toggles for DNS and HTTPS) and produces all nine resources with correct dependency ordering. Deployment takes roughly 15-20 minutes — compared to 4-8 hours of manual setup.

<!-- SCREENSHOT: Infra Chart detail page
  Page: /resource/infra-hub/infra-chart/[infraChartId]
  Action: Show a chart with template files and parameters visible
  Focus: The template YAML viewer and parameter list
  Alt: Infra Chart detail page showing template YAML files and a list of configurable parameters
-->

## Platform Charts and Custom Charts

Planton provides a set of curated charts for common infrastructure patterns, shipped with every [plantonhq/planton](https://github.com/plantonhq/planton) release. These include AWS ECS environments, EKS clusters, Pulumi state backends, and Terraform/OpenTofu state backends.

Organizations can also create custom charts tailored to their specific standards. The process is straightforward: follow the same directory structure (Chart.yaml, values.yaml, templates/), encode your organization's best practices into the templates, and publish the chart to Planton. Custom charts appear alongside platform charts in the catalog.

Common customizations include enforcing tagging policies, requiring HTTPS in production environments, setting minimum replica counts, and restricting VPC CIDR ranges to organizational standards.

<!-- SCREENSHOT: Infra Chart list
  Page: /resource/infra-hub/infra-chart
  Action: Show the list of available charts with platform and custom charts visible
  Focus: The chart list with names, descriptions, and ownership indicators
  Alt: Infra Chart list showing platform-provided and custom charts with descriptions
-->

## Charts vs. Direct Cloud Resources

Planton supports both deploying individual Cloud Resources and deploying bundled charts. Choose based on the situation:

| Scenario | Recommended Approach |
|----------|---------------------|
| Complete environment (dev, staging, production) | Infra Chart — coordinated deployment with dependency handling |
| Single resource (one S3 bucket, one DNS record) | Direct Cloud Resource — no orchestration overhead |
| Standardized patterns across teams | Infra Chart — encode best practices in templates |
| Quick experiments | Direct Cloud Resource — fast iteration, easy to delete |
| Complex multi-resource dependencies | Infra Chart — automatic DAG-based ordering |

You can start with a chart for initial deployment, then manage individual Cloud Resources afterward for incremental changes.

## Using the CLI

```bash
# Ground yourself before writing templates: the full schema reference for any
# resource kind or platform API, fully offline (kubectl-explain style)
planton explain aws-vpc
planton explain infra-chart

# Drill into one field with a dotted path of the exact YAML keys
planton explain aws-vpc.spec.instanceTenancy

# List every kind the CLI can explain (platform APIs + cloud resource kinds)
planton explain --list

# Build a chart from a local directory (renders and validates, reports issues)
planton chart build ./my-chart

# Build and print the combined rendered template to stdout
planton chart build ./my-chart --show

# Machine-readable report for CI pipelines and agents (exit code 0 = valid,
# 1 = chart has build errors, 2 = the check could not run)
planton chart build ./my-chart --output-format json

# Build a variant without editing the chart: override params for this one
# build (repeatable; the value is parsed as YAML, like values.yaml)
planton chart build ./my-chart --set dns_enabled=true --set region=eu-west-1

# Publish a chart to the organization's chart registry
planton chart publish ./my-chart

# List all charts available for the organization
planton chart list

# Create an Infra Project from a chart (deploys immediately)
planton chart install my-project ./my-chart -f values.yaml
```

The `chart install` command is the primary entry point for deploying a chart. It creates an Infra Project with the provided parameter values and automatically triggers an Infra Pipeline.

## Related Documentation

- [Cloud Resources](/docs/infrastructure/cloud-resources) — The individual resources that charts compose
- [Cloud Resource Kinds](/docs/infrastructure/cloud-resource-kinds) — The catalog of resource types available in charts
- [Infra Projects](/docs/infrastructure/infra-projects) — Running instances of Infra Charts
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — How chart deployments are orchestrated
- [Open Source](/docs/infrastructure/open-source) — The open-source foundation that provides resource APIs and cross-resource dependencies
