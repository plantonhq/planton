---
title: "Infra Components"
description: "The fundamental unit of infrastructure in Planton — a deployed instance of any catalog kind, tracked through its full lifecycle."
icon: cloud
order: 20
tags:
  - Infrastructure
  - Infra Components
  - Infrastructure
---

# Infra Components

An Infra Component is a deployed instance of a catalog kind — an AWS VPC, a GCP Cloud SQL database, an Azure AKS cluster, a Kubernetes deployment. Every piece of infrastructure Infra Hub manages exists as an Infra Component.

## Why Infra Components Exist

Managing infrastructure across multiple cloud providers typically means juggling different APIs, CLIs, and consoles — each with its own conventions for creating, updating, and destroying resources. A VPC in AWS has nothing in common with a GKE cluster in GCP, at least at the API level.

Infra Components provide a single, unified interface for all of it. Regardless of which provider or catalog kind you are working with, you define what you want, Planton provisions it through the appropriate Infrastructure-as-Code engine (Pulumi, Terraform, or OpenTofu), and the resource is tracked through its full lifecycle — from creation through updates to eventual teardown. One API, one CLI, one web console for infrastructure across 8 cloud providers and 700+ resource types.

## What an Infra Component Represents

When you create an Infra Component, you are declaring a piece of infrastructure: "I want an AWS RDS instance in my production environment with these settings." Planton takes that declaration and turns it into real cloud infrastructure by:

1. Matching the resource type to the appropriate IaC module
2. Resolving the provider credentials for your target environment
3. Running an [Infra Job](/docs/infrastructure/infra-jobs) that executes the provisioning
4. Tracking the resource's state and outputs

The resource definition follows a declarative model — you describe the desired state, and Planton handles the execution. Changes to the configuration trigger new Infra Jobs that reconcile the actual infrastructure with your updated definition.

## Two Deployment Paths

### Direct Deployment

Create and manage individual Infra Components directly. You define the resource configuration, submit it, and an Infra Job provisions it. This is the simplest path for standalone resources — a single database, a DNS zone, a storage bucket.

### Orchestrated Deployment

Deploy Infra Components as part of an [Infra Stack](/docs/infrastructure/infra-stacks), which coordinates multiple resources through an [Infra Pipeline](/docs/infrastructure/infra-pipelines). The pipeline resolves dependencies between resources, executes Infra Jobs in the correct order, and provides a unified view of the entire deployment. This is the path for multi-resource environments where resources depend on each other — for example, a VPC that must exist before the database that lives inside it.

## Lifecycle Operations

Infra Components support five primary operations:

- **Create** — Provision new infrastructure. Submitting an Infra Component definition triggers an Infra Job that creates the actual cloud infrastructure.
- **Update** — Modify the configuration. Changing an Infra Component triggers a new Infra Job that reconciles the infrastructure with the updated definition.
- **Destroy** — Tear down the cloud infrastructure. The Infra Job removes the actual resources from the cloud provider. The Infra Component record remains in Planton for audit purposes.
- **Import** — Adopt an existing provider resource into an Infra Component's IaC state without recreating it. The actual infrastructure is not modified — only the state file is updated. Use import when a provider resource was created outside of Planton (manually, by a provider, or through another tool) and you want Planton to manage it going forward. See [Importing Resources](/docs/infrastructure/importing-resources) for the full guide.
- **Purge** — Destroy the infrastructure and delete the Infra Component record from Planton in a single operation. Use this for full cleanup.

The distinction between destroy and purge matters for compliance and auditing. Destroy leaves a record of what existed and when it was torn down. Purge removes all traces.

A destroy reads only what it destroys. A `$secret/` or `$var/` reference the resource was deployed with that has since been deleted does not stop the destroy: the job names the references it could not read and tears the resource down. And a destroy that leaves the IaC state empty removes that state -- the state object in your bucket, or a Pulumi stack with its backups -- and the job says so (`state removed: <key>`); a Terraform Cloud workspace or a Pulumi Cloud stack is kept, holding its history, and a backend that refuses the delete keeps the state, with its reason.

<!-- SCREENSHOT: Infra Component detail page
  Page: /[org]/infra-component/[env]/[resourceKind]/[resourceName]
  Action: Show a deployed Infra Component with status and spec visible
  Focus: Full page showing resource metadata, spec summary, and Infra Job status
  Alt: Infra Component detail page showing a deployed AWS VPC with its configuration and latest Infra Job status
-->

## Provider Credentials

Each Infra Component is associated with a cloud provider and a connection that supplies the credentials for that provider. The provider is determined by the resource type — an AWS VPC uses an AWS connection, a GKE cluster uses a GCP connection.

If you do not specify a connection when creating a resource, Planton resolves the default connection for that provider and environment. Defaults can be set at the organization level (applies everywhere) or per-environment (overrides the organization default). See [Default Connections](/docs/connections/default-connections) for the full resolution logic.

Provider and IaC provisioner settings are fixed at creation time — changing them mid-lifecycle would require destroying and recreating the resource.

## Using the Web Console

### Creating an Infra Component

The Infra Components tab in Infra Hub guides you through a three-step process:

1. **Create an Environment** — All Infra Components belong to an environment (dev, staging, production).
2. **Connect a Provider** — Bring your cloud account or Kubernetes cluster into Planton through [Connections](/docs/connections).
3. **Select, Configure, and Deploy** — Browse the [Infra Catalog](/docs/infrastructure/catalog-kinds) to find the catalog kind you need, configure it, and deploy.

<!-- SCREENSHOT: Infra Component creation flow
  Page: /resource/infra-hub/infra-component/[provider]/[resource-kind]/create
  Action: Show the creation form with spec fields visible
  Focus: The form fields and provider-specific configuration
  Alt: Infra Component creation form for an AWS RDS instance showing spec configuration fields
-->

### Viewing Infra Components

The Infra Components tab offers two views:

- **List view** — A table showing environment, resource type, name, creator, and actions. Useful for scanning and filtering.
- **Grid view** — A visual canvas showing resources grouped by environment and type, with dependency connections between them. Useful for understanding relationships.

The detail panel for any Infra Component shows three tabs:

- **Configuration** — The resource's current settings
- **Versions** — History of configuration changes
- **Infra Jobs** — The IaC execution history for this resource

## Using the CLI

```bash
# Create an Infra Component from a YAML manifest
planton create -f manifest.yaml

# Get an Infra Component by ID
planton get infra-component <infra-component-id>

# Read one back as its manifest (kind, metadata, spec, status with outputs)
planton get AwsVpc production-vpc -o yaml

# The same resource inside the platform's InfraComponent wrapper
planton get AwsVpc production-vpc -o yaml --envelope

# Compare a local manifest with what is stored, before applying it
planton diff -f vpc.yaml

# List Infra Components
planton list infra-component

# Destroy the infrastructure (runs an Infra Job, keeps the record)
planton destroy <infra-component-id>

# Destroy infrastructure and delete the record
planton purge <infra-component-id>

# List every catalog kind
planton explain --list

# View the resource's infrastructure inputs
planton infra component iac-input <infra-component-id>

# Manage resource locks
planton infra component list-locks <infra-component-id>
planton infra component remove-locks <infra-component-id>

# Import an existing resource into IaC state (Pulumi)
planton pulumi import <infra-component> --type <type> --name <name> --id <provider-id>

# Import an existing resource into IaC state (Terraform / OpenTofu)
planton terraform import <infra-component> --address <address> --id <provider-id>
planton tofu import <infra-component> --address <address> --id <provider-id>
```

`planton get <Kind> <name> -o yaml` prints the resource as you write it, in the same YAML as every manifest: camelCase keys, with its outputs under `status.outputs` (for example `status.outputs.vpcId`). You can edit that output and apply it again. `-o json` prints proto field names (`status.outputs.vpc_id`).

`planton diff -f <manifest>` compares one local manifest with the record it would apply to, leaving status and the platform's own stamps out. It prints a unified diff and exits 1 when they differ, and exits 0 when applying would change nothing; `-o json` gives the list of changed fields. When nothing is stored under the manifest's name it prints **Nothing Stored Yet** and exits 3. It takes one manifest per file.

## Related Documentation

- [Catalog Kinds](/docs/infrastructure/catalog-kinds) — The catalog of available resource types
- [Infra Jobs](/docs/infrastructure/infra-jobs) — How infrastructure changes are executed
- [Infra Stacks](/docs/infrastructure/infra-stacks) — Orchestrated multi-resource deployments
- [Flow Control](/docs/infrastructure/flow-control) — Governance policies for deployment workflows
- [Importing Resources](/docs/infrastructure/importing-resources) — Bringing existing cloud infrastructure under management
- [Default Connections](/docs/connections/default-connections) — How provider credentials are resolved
