# Attach Service Project

Attaches an application project to the organization's Shared VPC host,
both by reference, so the chart enables the host before attaching and
detaches before disabling.

## What it configures

- `hostProjectId` — a reference to the `GcpSharedVpcHost`'s
  `host_project_id` output (the host KIND, which is what orders the chart).
- `serviceProjectId` — a reference to the application `GcpProject`.
- `deletionPolicy` left empty — the project is detached on destroy.

## Adjust before deploying

- **`hostProjectId.valueFrom.name`** and **`serviceProjectId.valueFrom.name`**
  — your host's and project's manifest names, or literal project IDs.
- **Grant `roles/compute.networkUser`** on the host's subnetworks to the
  service project's deployers and service agents with `GcpProjectIamMember`;
  the attachment alone does not let anything deploy.

## When to choose something else

For a project whose workloads must survive the chart that attached it, use
the **Attach Abandon On Destroy** preset.
