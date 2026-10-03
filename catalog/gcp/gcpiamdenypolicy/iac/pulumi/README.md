# GCP IAM Deny Policy - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying an IAM deny policy using Planton's `GcpIamDenyPolicy` API. The module is written in Go and creates `iam.DenyPolicy` — rules that BLOCK principals from using specific permissions regardless of any role grants they hold. Deny always outranks allow, which makes deny policies the guardrail layer.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **IAM permissions**: the deploying principal's permissions are listed in [`../permissions.yaml`](../permissions.yaml) — they must be granted at the ORGANIZATION level, even for project-attached policies
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator
    ├── deny_policy.go # Policy creation + parent encoding
    ├── locals.go      # Resolved resource + derived values
    └── outputs.go     # Output constants
```

## Quick Start

```bash
cd iac/pulumi
pulumi stack init dev
```

Provide a `iac-input.yaml`:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpIamDenyPolicy
  metadata:
    name: guard-secrets
  spec:
    rules:
      - description: Nobody reads unseal keys directly
        deny_rule:
          denied_principals:
            - principalSet://goog/public:all
          denied_permissions:
            - secretmanager.googleapis.com/versions.access
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpIamDenyPolicyIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpIamDenyPolicy` spec (parent, rules, exceptions, conditions) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `policy_name` | string | `{url-encoded-parent}/{policy_name}` — the policy's identifier |
| `etag` | string | The policy's current etag |

## Behavior Notes

- **The module renders the URL-encoded parent**: GCP's API identifies the attach point by its URL-encoded full resource name (e.g. `cloudresourcemanager.googleapis.com%2Fprojects%2Fmy-project`); the spec's typed parent message (project/folder/organization) keeps that encoding out of manifests.
- **An empty parent falls back to the provider's default project**, read from the provider's resolved configuration — gated to that one case so plans that name their attach point run credential-free.
- **No API enablement**: deny policies ride the always-on IAM v2 surface.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
