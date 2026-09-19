# GCP Shared VPC Host - Pulumi Module

## Overview

This directory contains the Pulumi implementation for enabling one project as a Shared VPC host using Planton's `GcpSharedVpcHost` API. The module is written in Go and creates `compute.SharedVPCHostProject` — a flag on the project — through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/compute.xpnAdmin` on the ORGANIZATION (or the folder above the project) — Shared VPC administration is organization-level; a project-level role is not enough

## Directory Structure

```
iac/pulumi/
├── main.go                 # Pulumi program entry point
├── Pulumi.yaml             # Pulumi project configuration
├── README.md               # This file
└── module/
    ├── main.go             # Module coordinator
    ├── shared_vpc_host.go  # The host flag
    ├── locals.go           # Resolved target
    └── outputs.go          # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Enables the spec's `project_id` (a `GcpProject`'s `project_id` output or a literal) as a Shared VPC host. When the spec names no project, the provider's default project is read from the client config (`organizations.GetClientConfig`) — this resource requires an explicit project argument, unlike most GCP resources.
- Sends the deletion policy only when set.
- Exports `host_project_id`, the resolved project.

## Parity with Terraform

Both engines send the same argument and produce the same output; the Terraform module's `google_client_config` data source is the counterpart of the client-config read here. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **The project is immutable**: a different host is a new resource.
- **Disabling a host fails while service projects are attached** — destroy the `GcpSharedVpcServiceProject` resources first (a chart's dependency order does this when they reference the host).
