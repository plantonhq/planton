# GCP Organization Policy Custom Constraint - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying one custom organization-policy constraint using Planton's `GcpOrgPolicyCustomConstraint` API. The module is written in Go and creates `orgpolicy.CustomConstraint` — an organization-defined rule that `GcpOrgPolicy` resources enforce by reference — through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/orgpolicy.policyAdmin` on the organization

## Directory Structure

```
iac/pulumi/
├── main.go                  # Pulumi program entry point
├── Pulumi.yaml              # Pulumi project configuration
├── README.md                # This file
└── module/
    ├── main.go              # Module coordinator
    ├── custom_constraint.go # The constraint
    ├── locals.go            # The custom.-prefixed name and the parent
    └── outputs.go           # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Names the constraint `custom.{constraint_name}` (the spec's bare name, defaulting to `metadata.name`) under `organizations/{organization_id}`.
- Sends the resource types, method types, condition, and action; optional display name, description, and deletion policy only when set.
- Exports `name` (the full resource name), `constraint` (the `custom.{name}` handle a policy references), and `update_time`.

## Parity with Terraform

Both engines send the same arguments and produce the same three outputs. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **A definition, not an enforcement** — nothing happens until a `GcpOrgPolicy` at some scope references this constraint.
- **Immutable identity** — the name, the organization, and the resource types recreate the constraint when changed; the rule itself (condition, action, methods) updates in place.
