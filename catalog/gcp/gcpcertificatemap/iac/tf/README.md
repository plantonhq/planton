# GCP Certificate Map - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Certificate Manager certificate map using Planton's `GcpCertificateMap` API. The module creates `google_certificate_manager_certificate_map` plus a `google_certificate_manager_certificate_map_entry` per spec entry — the hostname-to-certificate routing table an external HTTPS load balancer consults at TLS handshake time.

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Certificate Manager API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpCertificateMap`) — never hand-edited |
| `locals.tf` | Project fallback + name derivation + label merge + map_id assembly |
| `main.tf` | API enablement + map + entry fan-out |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `map_name` | `name` (map) | Defaults to metadata.name; ForceNew |
| `description` | `description` | |
| `labels` | `labels` | Platform attribution labels win; applied to map and entries |
| `entries[].entry_name` | `name` (entry) | ForceNew |
| `entries[].hostname` / `entries[].matcher` | `hostname` / `matcher` | Exactly one (provider ExactlyOneOf, spec-enforced); both ForceNew |
| `entries[].certificates[]` | `certificates` | 1–15; each references a GcpCertManagerCert's `certificate_id` output; the MUTABLE rotation surface |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` on map and every entry | One spec lever, every resource |

Certificate maps are GLOBAL (no location argument by API design).

## Outputs

| Output | Description |
|---|---|
| `map_id` | Full resource name (`projects/{p}/locations/global/certificateMaps/{name}`) |
| `map_uri` | The `//certificatemanager.googleapis.com/...` form a GcpTargetHttpsProxy's `certificate_map` argument consumes |
| `map_name` | The short map name |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
