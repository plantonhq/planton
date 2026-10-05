# GCP Certificate Map - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Certificate Manager certificate map using Planton's `GcpCertificateMap` API. The module is written in Go and creates `certificatemanager.CertificateMapResource` plus a `certificatemanager.CertificateMapEntry` per spec entry — the hostname-to-certificate routing table an external HTTPS load balancer consults at TLS handshake time.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Certificate Manager API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── certificate_map.go     # Map + entry fan-out
    ├── locals.go              # Resolved resource + derived values + label merges
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `map_name` | `name` (map) | Defaults to metadata.name; ForceNew |
| `description` | `description` | |
| `labels` | `labels` | Spec labels merged with platform attribution labels (platform wins); applied to map and entries |
| `entries[].entry_name` | `name` (entry) | ForceNew |
| `entries[].hostname` / `entries[].matcher` | `hostname` / `matcher` | Exactly one (provider ExactlyOneOf, spec-enforced); both ForceNew |
| `entries[].certificates[]` | `certificates` | 1–15; each references a GcpCertManagerCert's `certificate_id` output; the MUTABLE rotation surface |
| `entries[].description`, `entries[].labels` | same names | |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` on map and every entry | One spec lever, every resource |

Certificate maps are GLOBAL (no location argument by API design). The
module also enables `certificatemanager.googleapis.com` on the target
project (`disable_on_destroy` false).

## Outputs

| Output | Description |
|---|---|
| `map_id` | Full resource name (`projects/{p}/locations/global/certificateMaps/{name}`) |
| `map_uri` | The `//certificatemanager.googleapis.com/...` form a GcpTargetHttpsProxy's `certificate_map` argument consumes |
| `map_name` | The short map name |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
