# GCP Firewall Rule - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP compute firewall rules using Planton's `GcpFirewallRule` API. The module is written in Go and leverages the Pulumi GCP provider to create `google_compute_firewall` resources.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with billing enabled and Compute Engine API already active
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: `roles/compute.securityAdmin` or `roles/compute.networkAdmin` on the target project

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── debug.sh          # Debug helper script
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Module coordinator
    ├── firewall.go   # Firewall resource creation
    ├── locals.go     # Local values and labels
    └── outputs.go    # Stack output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `stack-input.yaml` with the firewall rule specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1
  kind: GcpFirewallRule
  metadata:
    name: allow-web
  spec:
    project_id:
      value: my-gcp-project-123
    network:
      value: my-vpc
    rule_name: allow-http-https
    direction: INGRESS
    action: ALLOW
    rules:
      - protocol: tcp
        ports: ["80", "443"]
    source_ranges: ["0.0.0.0/0"]
    target_tags: ["web-server"]

providerConfig:
  gcpCredential:
    value: <base64-encoded-service-account-key>
```

### 3. Deploy

```bash
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output firewall_self_link
pulumi stack output firewall_name
pulumi stack output creation_timestamp
```

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `firewall_self_link` | string | Full self-link URI of the firewall rule |
| `firewall_name` | string | Name of the firewall rule in GCP |
| `creation_timestamp` | string | RFC 3339 creation timestamp |

## Makefile Targets

```bash
make deps          # Download and tidy dependencies
make vet           # Run go vet
make fmt           # Format code
make build         # Build (runs deps, vet, fmt)
make update-deps   # Update Planton dependencies to latest
```

## Related

- [Component README](../../README.md) — full API reference and usage guide
- [Architecture Overview](overview.md) — internal module design
- [Examples](../../examples.md) — comprehensive manifest examples
