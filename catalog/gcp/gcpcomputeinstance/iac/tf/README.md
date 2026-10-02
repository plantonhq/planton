# GcpComputeInstance - Terraform Module

This Terraform/OpenTofu module provisions a Compute Engine virtual machine (`google_compute_instance`). It is the Terraform-side implementation of the Planton `GcpComputeInstance` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Compute Engine API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. The instance name falls back to `metadata.name`, user labels are merged beneath the platform attribution labels (`planton-ai_*`), and SSH keys fold into the metadata `ssh-keys` key (newline-joined) — all identically to the Pulumi module.

**Spot is the sharp edge**: `provisioning_model = "SPOT"` requires the API's legacy preemptible flag and forbids automatic restart — both derived in locals so the spec's single switch stays honest. The boot disk honors the exactly-one-source contract (an existing disk attaches via `source`; image/snapshot create a fresh disk through `initialize_params`), attached disks are pre-existing `GcpComputeDisk` resources attached by reference, and `desired_status` starts/suspends/stops the VM in place. Updates that need a stop/restart (machine type, service account, ...) are performed only when `allow_stopping_for_update` is true.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin (`~> 8.3`; all fields GA on the pinned line)
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — ambient-project fallback, `instance_name` fallback to `metadata.name`, label merge, ssh-keys fold, Spot derivation (`preemptible` + `automatic_restart`)
- `main.tf` — API enablement + the instance (boot disk, attached/scratch disks, NICs, scheduling, shielded/confidential, GPUs, reservation affinity)
- `outputs.tf` — `instance_name`, `instance_id`, `self_link`, `internal_ip`, `external_ip`, `status`, `zone`, `machine_type`, `cpu_platform`

## Outputs

| Output | Description |
|--------|-------------|
| `instance_name` | Name of the instance |
| `instance_id` | Server-assigned unique numeric identifier |
| `self_link` | API self link |
| `internal_ip` | Primary internal IP (first interface) |
| `external_ip` | External IP of the first interface; `""` when the VM is private |
| `status` | Current status (RUNNING, TERMINATED, ...) |
| `zone` | Zone of the instance |
| `machine_type` | Machine type |
| `cpu_platform` | CPU platform the instance landed on |
