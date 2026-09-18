# GCP Compute MIG - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Compute Engine Managed Instance Group using Planton's `GcpComputeMig` API. The module creates the group's whole reference-linked stack: the instance TEMPLATE (`google_compute_instance_template` zonal / `google_compute_region_instance_template` regional), the GROUP MANAGER (`google_compute_instance_group_manager` / `google_compute_region_instance_group_manager`), an optional AUTOSCALER (`google_compute_autoscaler` / `google_compute_region_autoscaler`), stateful PER-INSTANCE CONFIGS (`google_compute_per_instance_config` / `google_compute_region_per_instance_config`), and queued RESIZE REQUESTS (`google_compute_resize_request` / `google_compute_region_resize_request`). Zonal vs regional selection follows the spec's `zone` XOR `region` selector via one `is_regional` local with inverted `count`/`for_each` gates.

## Prerequisites

1. **OpenTofu** (or Terraform) installed
2. **GCP Project** with the Compute Engine API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/tf/
├── main.tf        # All ten resources (count/for_each-gated zonal + regional pairs)
├── locals.tf      # Names, label merge, location selector, rotation prefix
├── variables.tf   # GENERATED from the spec proto (planton tofu generate-variables)
├── outputs.tf     # one(concat()) branch-independent outputs
├── provider.tf    # google ~> 8.3 pin
├── backend.tf     # local backend for direct runs
└── README.md      # This file
```

## The reference-link contract (how the layers wire)

- **Group manager → template**: the zonal group manager's `version`
  block references the template's `self_link_unique` — the link carrying
  the `?uniqueId=` the provider's diff suppression keys on, so a
  template ROTATION genuinely rolls the version. The regional template
  has no unique variant; the regional group references its plain
  `self_link`, and rotation is carried by the `name_prefix` change
  producing a new link.
- **Autoscaler → group manager**: `target` takes the manager resource's
  `self_link`.
- **Per-instance configs / resize requests → group manager**: referenced
  by the manager resource's `name` (the provider's expected form for
  these resource pairs).

## Template rotation (the immutability story)

Instance templates cannot be modified in place (labels excepted). The
module names templates with `name_prefix` = `"<mig-name>-"` (capped at
37 characters so the provider uses its readable timestamp naming) plus
`lifecycle { create_before_destroy = true }` — the canonical provider
rotation pattern: a template change creates the NEW template first, the
group manager repoints, then the old template is deleted. How running
VMs pick the new template up is the spec's `update_policy` (PROACTIVE
rolls automatically; OPPORTUNISTIC waits).

## Notable mappings and behaviors

- `template.labels` merge with the platform attribution labels
  (platform wins) — the only template surface GCP mutates in place.
- `template.startup_script` maps to `metadata_startup_script` (distinct
  from user metadata; re-runs on every boot).
- Spot semantics: `template.scheduling.provisioning_model = "SPOT"`
  derives the API's legacy `preemptible` flag and forces
  `automatic_restart` off (FLEX_START likewise) — identically to the
  Pulumi module.
- Explicit-send posture: disk `auto_delete` (default true) and the
  shielded-VM booleans are always sent, so a spec transition to the
  non-default value reaches the engine.
- `versions` empty = one default version on this kind's own template; an
  entry's empty `template_self_link` resolves to the own template, a
  non-empty value pins an external canary template URL.
- `resize_requests[].requested_run_duration_seconds` renders as the
  provider's STRING `seconds` — byte-identical to the Pulumi module.
- `deletion_policy` lands on every resource that carries it (manager,
  autoscaler, per-instance configs, resize requests, and the REGIONAL
  template); the ZONAL template has no deletion policy in the provider —
  it is always deleted on destroy.
- `workload_identity_config` (a SPIFFE ID per VM, optional X.509
  certificates) and `scheduling.host_error_timeout_seconds` land on both
  templates; the identity block is emitted only when the spec sets it and
  the timeout only when set (Compute Engine's default recovery otherwise).
  Template `name`/`name_prefix` are module-internal (the rotation
  mechanism), and CSEK raw-key encryption arms are deliberately not
  modeled (secure-by-default — use CMEK).

## Outputs

| Output | Description |
|---|---|
| `instance_group` | Full URL of the group's instance group — the LB backend handle |
| `self_link` | Self link of the instance group MANAGER |
| `current_template_self_link` | The active template's link (unique form on zonal groups) |
| `mig_name` | The group name in GCP |
| `location` | The group's zone or region |

## Usage

```bash
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```
