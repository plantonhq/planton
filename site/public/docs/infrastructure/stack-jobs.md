---
title: "Stack Jobs"
description: "The atomic execution unit that provisions, updates, and destroys cloud infrastructure — running IaC operations with full visibility and control."
icon: pipeline
order: 50
tags:
  - Infrastructure
  - Stack Jobs
  - Infrastructure
---

# Stack Jobs

A Stack Job is the execution unit behind every infrastructure change in Planton. When you create a Cloud Resource, update its configuration, or destroy it, a Stack Job runs the Infrastructure-as-Code operations that make the change real — initializing the module, refreshing state, previewing changes, and applying them. Every step is tracked, every operation is logged, and the entire execution is preserved for audit.

## Why Stack Jobs Exist

Running infrastructure automation manually is fragile. You start with `terraform apply` (or `tofu apply`), then realize you should preview first, then add a refresh to catch drift, then add approval gates for production — and before long you are maintaining custom CI pipelines just to deploy a database safely.

Stack Jobs codify this workflow into the platform. The right operations run in the right order, every time. Credentials are resolved automatically. Flow control policies enforce governance without requiring manual intervention. And every execution — successful or failed — is recorded with full logs and resource diffs for troubleshooting and compliance.

## How a Stack Job Executes

Every Stack Job follows a defined sequence of operations. Which operations run depends on what the job is doing.

### For Creating or Updating Infrastructure

1. **Initialize** — Set up the IaC module, download providers, configure the state backend.
2. **Refresh** — Synchronize the IaC state file with the actual state of cloud resources. This catches drift — changes made outside of Planton, such as manual edits in the cloud console.
3. **Preview** — Generate a plan showing what will change. Additions, modifications, and deletions are displayed before anything happens.
4. **Apply** — Execute the changes. Create new resources, modify existing ones, update configurations.

### For Destroying Infrastructure

1. **Initialize** — Same module and state backend setup.
2. **Refresh** — Synchronize with actual cloud state.
3. **Preview** — Show what will be destroyed.
4. **Destroy** — Remove the resources from the cloud provider.

### For Importing Existing Resources

1. **Initialize** — Same module and state backend setup.
2. **Import** — Read the resource from the cloud provider and write it to the IaC state file. The actual infrastructure is not modified.
3. **Refresh** — Confirm the imported resource's state matches what exists on the provider.

Import Stack Jobs do not include preview or apply steps. The import operation writes to state only. See [Importing Resources](/docs/infrastructure/importing-resources) for the full guide, including provisioner-specific identifiers and CLI commands.

[Flow control policies](/docs/infrastructure/flow-control) can modify this sequence. The refresh step can be skipped for speed. The preview step can be made mandatory. A pause can be inserted between preview and apply to require manual approval before changes take effect.

<!-- SCREENSHOT: Stack Job detail page
  Page: /resource/infra-hub/stack-job/[stackJobId]
  Action: Show a Stack Job in progress with operation steps visible
  Focus: The operation steps with status indicators and expandable logs
  Alt: Stack Job detail page showing initialize completed, refresh in progress, preview and apply pending
-->

## What Gets Resolved Before Execution

Before a Stack Job starts running, the platform resolves four things automatically.

### IaC Module

The Pulumi, Terraform, or OpenTofu module that contains the provisioning logic for the resource type. If the Cloud Resource specifies a custom module, that module is used. Otherwise, the platform looks up the registered module for the resource type — first in the organization's module registry, then in the platform-level registry.

### Provider Credentials

The cloud provider credentials needed to provision the infrastructure — an AWS connection, a GCP connection, or whatever the resource type requires. These are resolved from the Cloud Resource's configuration, which specifies the connection to use. See [Connections](/docs/connections) for how connections are managed.

### State Backend

Where the IaC state file is stored. Each Cloud Resource gets a dedicated state file, stored in a backend (such as an S3 bucket or Pulumi Cloud) configured at the organization level. This ensures that one resource's state operations never interfere with another's.

### Flow Control Policy

The governance rules for this execution — whether manual approval is required, whether preview runs before apply, whether to pause between steps. Resolved from the most specific policy that applies: a policy targeting this specific resource, then the environment, then the organization, then the platform default. See [Flow Control](/docs/infrastructure/flow-control) for the full resolution hierarchy.

If any of these four cannot be resolved — missing credentials, no registered module, no state backend configured — the Stack Job fails its preflight check and does not execute. The preflight report identifies exactly what is missing.

## Two Deployment Paths

### Direct

When you create, update, or destroy a single Cloud Resource, one Stack Job runs for that resource. This is the simplest path — one resource, one job, one execution.

### Orchestrated

When an [Infra Pipeline](/docs/infrastructure/infra-pipelines) deploys multiple Cloud Resources, each resource gets its own Stack Job. The pipeline coordinates execution order based on the dependency graph — ensuring that a VPC's Stack Job completes before the database that depends on it starts.

## Monitoring Progress

### Web Console

The Stack Job detail page shows real-time progress as the job executes. Each operation (initialize, refresh, preview, apply/destroy) has its own status indicator — queued, running, succeeded, or failed. Expanding an operation reveals the full logs, including resource-level details showing what is being created, modified, or deleted.

The Cloud Resource detail page includes a Stack Jobs tab listing all jobs that have run for that resource, with timestamps, operation types, and results.

### CLI

Stream logs from a running Stack Job:

```bash
planton stack-job stream-progress-events <stack-job-id>
```

This command (aliased as `logs`) streams progress events in real time, including resource-level changes and operation transitions.

## Controlling Execution

### Pausing and Resuming

When a flow control policy requires manual approval — either before execution starts or between preview and apply — the Stack Job pauses and waits. Resume it from the web console or CLI:

```bash
planton stack-job resume <stack-job-id>
```

### Cancelling

Cancel a running Stack Job to stop execution. The currently in-flight IaC operation completes to avoid leaving resources in an inconsistent state, then remaining operations are skipped:

```bash
planton stack-job cancel <stack-job-id>
```

### Re-running

Re-run a completed or failed Stack Job to repeat the same operation. Useful after fixing external issues like quota limits or permission errors:

```bash
planton stack-job rerun <stack-job-id>
```

## Preflight Checks

Before committing to execution, verify that all four essentials are in place for a given resource type and environment:

```bash
planton stack-job preflight-checks --cloud-resource-kind <kind>
```

The report shows whether the IaC module, provider credentials, state backend, and flow control policy can all be resolved. This is useful when setting up a new environment or debugging why a Stack Job failed to start.

## Using the CLI

```bash
# Create a Stack Job for a Cloud Resource
planton stack-job create-stack-job <cloud-resource-id> --operation update --tail

# List Stack Jobs for a resource
planton stack-job list <cloud-resource-id>

# Stream logs from a running job (alias: logs)
planton stack-job stream-progress-events <stack-job-id>

# Cancel a running job
planton stack-job cancel <stack-job-id>

# Resume a paused job
planton stack-job resume <stack-job-id>

# Re-run a completed or failed job
planton stack-job rerun <stack-job-id>

# Run preflight checks
planton stack-job preflight-checks --cloud-resource-kind <kind>

# View the stack input for a completed job
planton stack-job stack-input <stack-job-id>
```

The `--operation` flag on `create-stack-job` accepts: `refresh`, `preview`, `update`, `destroy`. The default is `preview`.

Additional flags for `create-stack-job`, `resume`, and `rerun`:

- `--tail` / `-t` — Follow the job's progress events after creation
- `--show-stack-summary` — Display the resource summary at completion (default: on)
- `--show-stack-diff` — Display detailed diffs for changed resources
- `--show-stack-outputs` — Display the stack outputs at completion (default: on). A secret the resource generates shows as its `$secret/` reference; the value itself is kept in your secret store (see [Where Secrets Live](/docs/secrets/where-secrets-live#secrets-a-resource-generates))
- `--version-message` / `-m` — A description for the job, similar to a commit message

The `stack-job` command can also be invoked as `sj` for brevity.

## Related Documentation

- [Cloud Resources](/docs/infrastructure/cloud-resources) — The infrastructure that Stack Jobs provision
- [Importing Resources](/docs/infrastructure/importing-resources) — Adopting existing infrastructure into IaC state
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — Orchestrating multiple Stack Jobs across resources
- [Flow Control](/docs/infrastructure/flow-control) — Governance policies that control Stack Job execution
- [Connections](/docs/connections) — How provider credentials are managed and resolved
