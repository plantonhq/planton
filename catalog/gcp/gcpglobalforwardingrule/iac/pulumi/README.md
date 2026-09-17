# GCP Global Forwarding Rule - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine global forwarding rules using Planton's `GcpGlobalForwardingRule` API. The module is written in Go and creates `compute.GlobalForwardingRule`.

The forwarding rule is the VIP node of a global load balancer — it binds an IP address and port to a target proxy — and doubles as the Private Service Connect entry point.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with Compute Engine API available
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: any role carrying `compute.globalForwardingRules.*` on the target project
6. **A target proxy** — a `GcpTargetHttpsProxy` (or `GcpTargetHttpProxy`) whose `self_link` you wire in

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go                     # Module coordinator
    ├── global_forwarding_rule.go   # Rule creation and mapping
    ├── locals.go                   # Resolved resource + derived values (incl. the NONE→"" scheme mapping)
    └── outputs.go                  # Stack output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `stack-input.yaml` with the rule specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpGlobalForwardingRule
  metadata:
    name: web-frontend-443
  spec:
    project_id:
      value: my-gcp-project-123
    target:
      value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/targetHttpsProxies/web-https-frontend
    port_range: "443"
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output ip_address
pulumi stack output self_link
```

## Inputs

The module consumes `GcpGlobalForwardingRuleStackInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpGlobalForwardingRule` spec (target proxy, VIP, scheme, PSC extras) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `ip_address` | string | The VIP — the value DNS records point at |
| `self_link` | string | Self-link URI of the rule |
| `forwarding_rule_name` | string | Name of the rule in GCP |
| `forwarding_rule_id` | string | Server-assigned numeric ID |
| `psc_connection_id` | string | PSC connection id (PSC frontends only) |
| `psc_connection_status` | string | PSC connection status (PSC frontends only) |

## Behavior Notes

- **Mutability**: only `target` (via `setTarget`) and `labels` update in place — everything else recreates the rule, so bind a reserved static IP for production.
- **The PSC sentinel**: spec scheme `NONE` is sent to the API as an EMPTY scheme string — Private Service Connect's form.
- **The scheme is always sent**: an unset scheme is sent as `EXTERNAL` (the spec's default, the classic global external ALB) rather than omitted. The provider's own default is `EXTERNAL_MANAGED` and the scheme is immutable, so leaving the choice to the provider would replace an existing classic frontend on its next apply. The Terraform module does the same.
- **Ambient project**: an empty `project_id` falls back to the provider's default project.
- **API enablement**: the module enables `compute.googleapis.com` before creating the rule (`disable_on_destroy=false`).

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
