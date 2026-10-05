# AwsSagemakerNotebookInstance — Kind Guide

Authored operational judgment for the SageMaker notebook instance
kind: the design decisions behind the spec's shape, and what to
know before running managed notebooks in production.

## Design decisions

- **The instance's AWS name derives from `metadata.name`**, and the
  folded lifecycle configuration rides a stable derived name
  (`<name>-lifecycle`) — no separate name fields to keep in sync.
- **The lifecycle configuration is folded into the instance.** A
  bootstrap script has no life of its own — it exists to shape this
  instance. Authors write plain shell in `on_create` / `on_start`; the
  modules base64-encode before sending (the API's transport format),
  so no manifest ever carries encoded blobs.
- **At least one script is required when `lifecycle_config` is set.**
  AWS accepts an empty lifecycle configuration, but it configures
  nothing — always authoring drift, so the spec rejects it.
- **VPC confinement is validated as a unit.** `direct_internet_access:
  Disabled` requires `subnet_id` and `security_group_ids` (AWS needs
  the VPC wiring to confine the notebook), and security groups are
  meaningless without a subnet — both enforced at manifest time.
- **`volume_size_gb` names the unit** and carries AWS's bounds
  (5–16384); the growth-only rule is behavioral, taught on the field
  and enforced by the provider.

## Running notebook instances in production

- **Budget minutes per change.** SageMaker requires a Stopped instance
  for most updates — the modules ride the provider's stop-update-start
  choreography, and the notebook is unavailable through it. Batch
  changes rather than trickling them.
- **Never shrink the volume.** Growing updates in place; shrinking
  replaces the instance and everything on the volume. Size generously
  — the storage is cheap relative to a rebuilt workspace.
- **Keep bootstrap scripts under five minutes.** They run as root at
  create/start, and a script that exceeds AWS's limit fails the
  instance start — push long installs to the background (`nohup … &`)
  or bake a custom environment instead.
- **Replacing script text beats clearing it.** The provider's update
  omits empty fields, so clearing a script in the spec does NOT clear
  it in AWS — replace the text (even with a no-op `#!/bin/bash`) when
  retiring behavior.
- **Know the replacement triggers.** `subnet_id`,
  `direct_internet_access`, and `platform_identifier` changes replace
  the instance — decide network posture and platform up front.
- **Lock down what users don't need.** `root_access: Disabled` keeps
  users out of root (lifecycle scripts still run as root);
  `imds_minimum_version: "2"` is the hardened metadata-service choice;
  prefer `notebook-al2-v3` or `notebook-al2023-v1` — the older
  platforms are deprecated, kept only for existing workloads.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
