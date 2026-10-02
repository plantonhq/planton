# GCP Compute MIG - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Compute Engine Managed Instance Group using Planton's `GcpComputeMig` API. The module is written in Go and creates the group's whole reference-linked stack: the instance TEMPLATE (`compute.InstanceTemplate` zonal / `compute.RegionInstanceTemplate` regional), the GROUP MANAGER (`compute.InstanceGroupManager` / `compute.RegionInstanceGroupManager`), an optional AUTOSCALER (`compute.Autoscaler` / `compute.RegionAutoscaler`), stateful PER-INSTANCE CONFIGS (`compute.PerInstanceConfig` / `compute.RegionPerInstanceConfig`), and queued RESIZE REQUESTS (`compute.ResizeRequest` / `compute.RegionResizeRequest`). Zonal vs regional selection follows the spec's `zone` XOR `region` selector.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Compute Engine API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                        # Pulumi program entry point
├── Pulumi.yaml                    # Pulumi project configuration
├── README.md                      # This file
└── module/
    ├── main.go                    # Module coordinator (create order + exports)
    ├── instance_template.go       # Zonal/regional template + rotation naming
    ├── instance_group_manager.go  # Zonal/regional group manager + versions
    ├── autoscaler.go              # Zonal/regional autoscaler (optional)
    ├── per_instance_config.go     # Stateful per-instance configs
    ├── resize_request.go          # Queued capacity requests
    ├── locals.go                  # Names, label merge, location selector
    └── outputs.go                 # Stack output constants
```

## The reference-link contract (how the layers wire)

- **Group manager → template**: the zonal group manager's version block
  references the template's `self_link_unique` — the link that carries
  the `?uniqueId=` the provider's diff suppression keys on, so a
  template ROTATION (new template) genuinely rolls the version. The
  regional template has no unique variant; the regional group references
  its plain `self_link`, and rotation is carried by the `name_prefix`
  change producing a new link.
- **Autoscaler → group manager**: `target` takes the manager's
  `self_link` output.
- **Per-instance configs / resize requests → group manager**: referenced
  by the manager's NAME output (the provider's expected form for these
  resource pairs).

## Template rotation (the immutability story)

Instance templates cannot be modified in place (labels excepted). The
module names templates with `name_prefix` = `"<mig-name>-"` (capped at
37 characters so the provider uses its readable timestamp naming) and
relies on Pulumi's default create-before-delete replacement: a template
change creates the NEW template first, repoints the group manager, then
deletes the old one — the group is never left referencing a deleted
template. How running VMs pick the new template up is the spec's
`update_policy` (PROACTIVE rolls automatically; OPPORTUNISTIC waits).

## How the module maps the spec (headline mappings)

| Spec field | Provider argument | Notes |
|---|---|---|
| `mig_name` | manager `name` | Defaults to metadata.name; seeds template prefix + autoscaler name |
| `zone` / `region` | resource family selector | Exactly one; picks zonal vs regional resources kind-wide |
| `base_instance_name` | `base_instance_name` | Defaults to the group name |
| `template.*` | template resource | Flat `disk` blocks (role fields on the disk), NICs, scheduling, shielded/confidential, accelerators |
| `template.labels` | template `labels` | Merged with platform attribution labels (platform wins) |
| `template.startup_script` | `metadata_startup_script` | Distinct from user metadata |
| `versions[]` | manager `version` blocks | Empty list = one default version on the kind's own template; `template_self_link` pins an external canary template |
| `target_size` | manager `target_size` | Mutually exclusive with `autoscaler` (spec CEL) |
| `auto_healing.health_check` | `auto_healing_policies.health_check` | References a GcpHealthCheck's `self_link` output |
| `autoscaler.*` | autoscaler resource | Created only when the message is set |
| `per_instance_configs[]` | per-instance config resources | Config name IS the instance name |
| `resize_requests[]` | resize request resources | `requested_run_duration_seconds` rendered as the provider's STRING seconds |
| `deletion_policy` | per-resource `deletion_policy` | Lands on manager, autoscaler, configs, requests, and the REGIONAL template; the ZONAL template carries none in the provider (always deleted) |

Spot semantics: `template.scheduling.provisioning_model = "SPOT"` derives
the API's legacy preemptible flag and forces automatic restart off —
identically to the Terraform module.

`workload_identity_config` (a SPIFFE ID per VM, optional X.509
certificates) and `scheduling.host_error_timeout_seconds` land on both
templates — the identity block only when the spec sets it, the timeout
only when set — identically to the Terraform module. Template
`name`/`name_prefix` are module-internal (the rotation mechanism), and
CSEK raw-key encryption arms are deliberately not modeled
(secure-by-default — use CMEK).

The module also enables `compute.googleapis.com` on the target project
(`disable_on_destroy` false — tearing down one group never disables
Compute Engine project-wide).

## Stack Outputs

| Output | Description |
|---|---|
| `instance_group` | Full URL of the group's instance group — the LB backend handle (a GcpBackendService backend's `group` takes exactly this) |
| `self_link` | Self link of the instance group MANAGER |
| `current_template_self_link` | The active template's link (unique form on zonal groups — changes on every rotation) |
| `mig_name` | The group name in GCP |
| `location` | The group's zone or region |

## Local development

`stack-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
