# GcpVertexAiNotebook - Pulumi Module

This Pulumi (Go) module provisions a Vertex AI Workbench instance (`workbench.Instance`) — a managed JupyterLab environment backed by a Compute Engine VM. It is the Pulumi-side implementation of the Planton `GcpVertexAiNotebook` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Notebooks API (`notebooks.googleapis.com`) and the Compute API (`compute.googleapis.com`, the Workbench VM is a Compute Engine instance) with `disable_on_destroy=false`, so a fresh project works first try and teardown never disables the APIs project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module.

**The instance name is always sent explicitly** (`spec.instance_name`, falling back to `metadata.name`): the create call derives the cloud-side instance ID from `name`, and engine auto-naming would otherwise append a random suffix that breaks the spec's naming contract.

The bridged provider's client-side `deletion_policy` knob is pinned to `DELETE` so destroy behavior is byte-identical to the Terraform module.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpvertexainotebook/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/locals.go` — instance-name resolution + labels merge
- `module/workbench_instance.go` — API enablement + the Workbench instance (full gce_setup surface) + outputs
- `module/outputs.go` — output constant names
