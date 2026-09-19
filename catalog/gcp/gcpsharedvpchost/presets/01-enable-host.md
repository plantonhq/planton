# Enable Host

Enables the organization's network project as the Shared VPC host, by
reference to the `GcpProject` that creates it, and protects the role
against destroy because every service project depends on it.

## What it configures

- `projectId` — a reference to the `GcpProject`'s `project_id` output. Leave
  it out entirely to enable the project the credentials are configured for.
- `deletionPolicy: PREVENT` — the host cannot be disabled by a destroy.

## Adjust before deploying

- **`projectId.valueFrom.name`** — your network project's manifest name, or
  a literal project ID.
- **The connection's identity** must hold `roles/compute.xpnAdmin` on the
  organization or a folder above the project.

## When to choose something else

This is the only shape a host has. Attach service projects with the
`GcpSharedVpcServiceProject` presets.
