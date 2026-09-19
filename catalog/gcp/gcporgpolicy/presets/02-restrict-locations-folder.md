# Restrict Locations (Folder)

A list constraint on a folder: every resource created in any project
beneath `production` must live in a US location. The policy is the new
root for this constraint (`inheritFromParent: false`), so nothing set
higher in the hierarchy widens it.

## What it configures

- `scope.folderId` — a reference to the `GcpFolder`; everything beneath
  inherits the rule.
- `constraint: gcp.resourceLocations` — a list constraint, so the rule uses
  `values`.
- `allowedValues: [in:us-locations]` — a Google value group; `us-central1`
  and friends work too.

## Adjust before deploying

- **`allowedValues`** — the locations or value groups your data-residency
  rule allows (`in:eu-locations`, `europe-west1`).
- **`inheritFromParent`** — set true to ADD to the parent's list instead of
  replacing it.

## When to choose something else

A yes/no guardrail takes the **Disable Serial Port (Project)** preset.
