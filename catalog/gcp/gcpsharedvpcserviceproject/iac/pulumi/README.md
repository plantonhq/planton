# GCP Shared VPC Service Project - Pulumi Module

## Overview

This directory contains the Pulumi implementation for attaching one service project to a Shared VPC host using Planton's `GcpSharedVpcServiceProject` API. The module is written in Go and creates `compute.SharedVPCServiceProject` through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/compute.xpnAdmin` on the organization (or the folder above both projects)

## Directory Structure

```
iac/pulumi/
├── main.go                            # Pulumi program entry point
├── Pulumi.yaml                        # Pulumi project configuration
├── README.md                          # This file
└── module/
    ├── main.go                        # Module coordinator
    ├── shared_vpc_service_project.go  # The attachment
    ├── locals.go                      # Resolved target
    └── outputs.go                     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Attaches the spec's `service_project_id` (a `GcpProject`) to the spec's `host_project_id` (a `GcpSharedVpcHost`'s `host_project_id` output).
- Sends the deletion policy only when set (`ABANDON` is the only accepted value: leave the attachment in place on destroy).
- Exports `service_project_id` and `host_project_id`.

## Parity with Terraform

Both engines send the same arguments and produce the same two outputs. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **Both projects are immutable**: moving to another host is a detach and an attach.
- **Detaching fails while resources in the service project still use a host subnetwork.**
- **Attaching is not enough by itself**: the service project's deployers and service agents still need `roles/compute.networkUser` on the host's subnetworks — declared with `GcpProjectIamMember` or subnetwork-level IAM, not here.
