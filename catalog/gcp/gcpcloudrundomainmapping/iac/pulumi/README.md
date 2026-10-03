# GCP Cloud Run Domain Mapping - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Run domain mapping using Planton's `GcpCloudRunDomainMapping` API. The module is written in Go and creates `cloudrun.DomainMapping` — a verified custom domain pointed directly at a Cloud Run service, with Cloud Run provisioning the TLS certificate in AUTOMATIC mode.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Cloud Run Admin API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **A verified domain**: the provisioning identity must have verified domain ownership (Search Console / `gcloud domains verify`) — GCP rejects the create otherwise
6. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs -- including why domain mappings need more than any Cloud Run role grants

## Directory Structure

```
iac/pulumi/
├── main.go                  # Pulumi program entry point
├── Pulumi.yaml              # Pulumi project configuration
├── README.md                # This file
└── module/
    ├── main.go              # Module coordinator
    ├── domain_mapping.go    # Domain mapping creation
    ├── locals.go            # Resolved resource + derived values
    └── outputs.go           # Output constants
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
  kind: GcpCloudRunDomainMapping
  metadata:
    name: app-domain
  spec:
    region: us-central1
    domain: app.example.com
    route:
      value: my-service
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpCloudRunDomainMappingIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpCloudRunDomainMapping` spec (domain, region, route, certificate mode) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `domain` | string | The mapped domain (the mapping's name in GCP) |
| `region` | string | GCP region the mapping lives in |
| `resource_records` | list | DNS records the domain's zone must publish (A/AAAA for a root domain, one CNAME for a subdomain) |
| `mapped_route_name` | string | The Cloud Run route (service) the mapping currently points to |

## Behavior Notes

- **The mapping is fully immutable**: every provider argument is create-only, so any spec change replaces the mapping. Replacement is cheap (free object, re-creates in seconds) with a brief serving gap while the certificate re-issues.
- **The domain must be verified first** — a one-time, out-of-band step per domain; no IaC resource performs it.
- **The record set is server-decided** (a root domain receives A/AAAA sets, a subdomain one CNAME), so `resource_records` exports as one structured list registered synchronously — never `ctx.Export` inside an Apply callback, which races the engine's output marshaling.
- **The required metadata namespace** falls back spec.namespace → spec project → the provider's resolved default project (client-config read gated to that last case).
- **API enablement**: the module enables `run.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
