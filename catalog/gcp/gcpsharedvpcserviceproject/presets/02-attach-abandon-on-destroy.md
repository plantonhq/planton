# Attach Abandon On Destroy

Attaches a long-lived platform project to the host by literal IDs and
keeps the attachment when this resource is destroyed — for a project whose
workloads must outlive the chart that attached it.

## What it configures

- `hostProjectId` and `serviceProjectId` — literal project IDs, for a host
  and project created outside this chart.
- `deletionPolicy: ABANDON` — destroy unmanages the attachment and leaves
  the project attached with every workload intact. This is the only
  alternative to detaching that Google's resource accepts.

## Adjust before deploying

- **The two project IDs** — yours.

## When to choose something else

When the chart owns the project's lifecycle, use the **Attach Service
Project** preset so destroy detaches cleanly (after the workloads that use
the shared subnetworks are gone).
