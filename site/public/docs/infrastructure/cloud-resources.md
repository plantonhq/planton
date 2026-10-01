---
title: "Cloud Resources"
description: "The fundamental unit of infrastructure in Planton — a deployed instance of any cloud component, tracked through its full lifecycle."
icon: cloud
order: 20
tags:
  - Infrastructure
  - Cloud Resources
  - Infrastructure
---

# Cloud Resources

A Cloud Resource is a deployed instance of a cloud component — an AWS VPC, a GCP Cloud SQL database, an Azure AKS cluster, a Kubernetes deployment. Every piece of infrastructure managed through Infrastructure exists as a Cloud Resource.

## Why Cloud Resources Exist

Managing infrastructure across multiple cloud providers typically means juggling different APIs, CLIs, and consoles — each with its own conventions for creating, updating, and destroying resources. A VPC in AWS has nothing in common with a GKE cluster in GCP, at least at the API level.

Cloud Resources provide a single, unified interface for all of it. Regardless of which provider or resource type you are working with, you define what you want, Planton provisions it through the appropriate Infrastructure-as-Code engine (Pulumi, Terraform, or OpenTofu), and the resource is tracked through its full lifecycle — from creation through updates to eventual teardown. One API, one CLI, one web console for infrastructure across 8 cloud providers and 700+ resource types.

## What a Cloud Resource Represents

When you create a Cloud Resource, you are declaring a piece of infrastructure: "I want an AWS RDS instance in my production environment with these settings." Planton takes that declaration and turns it into real cloud infrastructure by:

1. Matching the resource type to the appropriate IaC module
2. Resolving the provider credentials for your target environment
3. Running a [Stack Job](/docs/infrastructure/stack-jobs) that executes the provisioning
4. Tracking the resource's state and outputs

The resource definition follows a declarative model — you describe the desired state, and Planton handles the execution. Changes to the configuration trigger new Stack Jobs that reconcile the actual infrastructure with your updated definition.

## Two Deployment Paths

### Direct Deployment

Create and manage individual Cloud Resources directly. You define the resource configuration, submit it, and a Stack Job provisions it. This is the simplest path for standalone resources — a single database, a DNS zone, a storage bucket.

### Orchestrated Deployment

Deploy Cloud Resources as part of an [Infra Project](/docs/infrastructure/infra-projects), which coordinates multiple resources through an [Infra Pipeline](/docs/infrastructure/infra-pipelines). The pipeline resolves dependencies between resources, executes Stack Jobs in the correct order, and provides a unified view of the entire deployment. This is the path for multi-resource environments where resources depend on each other — for example, a VPC that must exist before the database that lives inside it.

## Lifecycle Operations

Cloud Resources support five primary operations:

- **Create** — Provision new infrastructure. Submitting a Cloud Resource definition triggers a Stack Job that creates the actual cloud infrastructure.
- **Update** — Modify the configuration. Changing a Cloud Resource triggers a new Stack Job that reconciles the infrastructure with the updated definition.
- **Destroy** — Tear down the cloud infrastructure. The Stack Job removes the actual resources from the cloud provider. The Cloud Resource record remains in Planton for audit purposes.
- **Import** — Adopt an existing cloud resource into Planton's IaC state without recreating it. The actual infrastructure is not modified — only the state file is updated. Use import when a cloud resource was created outside of Planton (manually, by a provider, or through another tool) and you want Planton to manage it going forward. See [Importing Resources](/docs/infrastructure/importing-resources) for the full guide.
- **Purge** — Destroy the infrastructure and delete the Cloud Resource record from Planton in a single operation. Use this for full cleanup.

The distinction between destroy and purge matters for compliance and auditing. Destroy leaves a record of what existed and when it was torn down. Purge removes all traces.

A destroy reads only what it destroys. A `$secret/` or `$var/` reference the resource was deployed with that has since been deleted does not stop the destroy: the job names the references it could not read and tears the resource down. And a destroy that empties the stack removes the state it left behind -- the state object in your bucket, or a Pulumi stack with its backups -- and the job says so (`state removed: <key>`); a Terraform Cloud workspace or a Pulumi Cloud stack is kept, holding its history, and a backend that refuses the delete keeps the state, with its reason.

<!-- SCREENSHOT: Cloud Resource detail page
  Page: /[org]/cloud-resource/[env]/[resourceKind]/[resourceName]
  Action: Show a deployed Cloud Resource with status and spec visible
  Focus: Full page showing resource metadata, spec summary, and stack job status
  Alt: Cloud Resource detail page showing a deployed AWS VPC with its configuration and latest Stack Job status
-->

## Provider Credentials

Each Cloud Resource is associated with a cloud provider and a connection that supplies the credentials for that provider. The provider is determined by the resource type — an AWS VPC uses an AWS connection, a GKE cluster uses a GCP connection.

If you do not specify a connection when creating a resource, Planton resolves the default connection for that provider and environment. Defaults can be set at the organization level (applies everywhere) or per-environment (overrides the organization default). See [Default Connections](/docs/connections/default-connections) for the full resolution logic.

Provider and IaC provisioner settings are fixed at creation time — changing them mid-lifecycle would require destroying and recreating the resource.

## Using the Web Console

### Creating a Cloud Resource

The Cloud Resources tab in Infra Hub guides you through a three-step process:

1. **Create an Environment** — All Cloud Resources belong to an environment (dev, staging, production).
2. **Connect a Provider** — Bring your cloud account or Kubernetes cluster into Planton through [Connections](/docs/connections).
3. **Select, Configure, and Deploy** — Browse the [Deployment Component catalog](/docs/infrastructure/cloud-resource-kinds) to find the resource type you need, configure it, and deploy.

<!-- SCREENSHOT: Cloud Resource creation flow
  Page: /resource/infra-hub/cloud-resource/[provider]/[resource-kind]/create
  Action: Show the creation form with spec fields visible
  Focus: The form fields and provider-specific configuration
  Alt: Cloud Resource creation form for an AWS RDS instance showing spec configuration fields
-->

### Viewing Cloud Resources

The Cloud Resources tab offers two views:

- **List view** — A table showing environment, resource type, name, creator, and actions. Useful for scanning and filtering.
- **Grid view** — A visual canvas showing resources grouped by environment and type, with dependency connections between them. Useful for understanding relationships.

The detail panel for any Cloud Resource shows three tabs:

- **Configuration** — The resource's current settings
- **Versions** — History of configuration changes
- **Stack Jobs** — The IaC execution history for this resource

## Using the CLI

```bash
# Create a Cloud Resource from a YAML manifest
planton create -f manifest.yaml

# Get a Cloud Resource by ID
planton get cloud-resource <cloud-resource-id>

# Read one back as its manifest (kind, metadata, spec, status with outputs)
planton get AwsVpc production-vpc -o yaml

# The same resource inside the platform's CloudResource wrapper
planton get AwsVpc production-vpc -o yaml --envelope

# Compare a local manifest with what is stored, before applying it
planton diff -f vpc.yaml

# List Cloud Resources
planton list cloud-resource

# Destroy the infrastructure (runs a Stack Job, keeps the record)
planton cloud-resource destroy <cloud-resource-id>

# Destroy infrastructure and delete the record
planton cloud-resource purge <cloud-resource-id>

# List all available resource types
planton cloud-resource registered-kinds

# View the resource's infrastructure inputs
planton cloud-resource stack-input <cloud-resource-id>

# Manage resource locks
planton cloud-resource list-locks <cloud-resource-id>
planton cloud-resource remove-locks <cloud-resource-id>

# Import an existing resource into IaC state (Pulumi)
planton pulumi import <cloud-resource> --type <type> --name <name> --id <provider-id>

# Import an existing resource into IaC state (Terraform / OpenTofu)
planton terraform import <cloud-resource> --address <address> --id <provider-id>
planton tofu import <cloud-resource> --address <address> --id <provider-id>
```

`planton get <Kind> <name> -o yaml` prints the resource as you write it, in the same YAML as every manifest: camelCase keys, with its outputs under `status.outputs` (for example `status.outputs.vpcId`). You can edit that output and apply it again. `-o json` prints proto field names (`status.outputs.vpc_id`).

`planton diff -f <manifest>` compares one local manifest with the record it would apply to, leaving status and the platform's own stamps out. It prints a unified diff and exits 1 when they differ, and exits 0 when applying would change nothing; `-o json` gives the list of changed fields. When nothing is stored under the manifest's name it prints **Nothing Stored Yet** and exits 3. It takes one manifest per file.

## Related Documentation

- [Cloud Resource Kinds](/docs/infrastructure/cloud-resource-kinds) — The catalog of available resource types
- [Stack Jobs](/docs/infrastructure/stack-jobs) — How infrastructure changes are executed
- [Infra Projects](/docs/infrastructure/infra-projects) — Orchestrated multi-resource deployments
- [Flow Control](/docs/infrastructure/flow-control) — Governance policies for deployment workflows
- [Importing Resources](/docs/infrastructure/importing-resources) — Bringing existing cloud infrastructure under management
- [Default Connections](/docs/connections/default-connections) — How provider credentials are resolved
