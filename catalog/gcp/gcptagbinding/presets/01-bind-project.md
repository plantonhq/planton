# Bind Project

Tags a project with `environment/prod`, both sides by reference. The
`GcpProject` reference resolves to the project NUMBER Google requires, so
no lookup runs.

## What it configures

- `tagValue` — a reference to the `GcpTagValue`'s `name` output.
- `parent.projectId` — a reference to the `GcpProject`'s `project_number`.
  Omit `parent` entirely to tag the project you are deploying into.

## Adjust before deploying

- **The two references** — your value and project manifest names. A literal
  project ID also works (looked up once at apply time), as does a number.

## When to choose something else

A folder takes the **Bind Folder** preset; a regional or zonal resource
takes **Bind Regional Resource**.
