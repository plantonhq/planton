# Org Deny SSH From Internet

The organization's baseline: two logged denies nobody in any project can
undo -- SSH and RDP from the internet -- followed by `goto_next` so every
other decision is left to the folders, the networks, and the projects.
Enforced on the organization itself and guarded against destroy.

## What it configures

- `parent.organizationId` and the association target — the policy lives on
  the organization and is enforced on all of it.
- Priorities `1000` and `1100` — `deny` tcp:22 and tcp:3389 from
  `0.0.0.0/0`, each with `enableLogging: true` so a blocked connection is
  explainable in Cloud Logging.
- Priority `2000` — `goto_next` on `all`, the delegation that lets the
  levels below decide everything the baseline does not.
- `deletionPolicy: PREVENT` — deleting a baseline silently reopens what it
  closed for every project.

## Adjust before deploying

- **`organizationId`** (both places) — your numeric organization ID from
  `gcloud organizations list`.
- **Ports** — add or remove administrative ports the organization never
  exposes publicly.

## When to choose something else

For a folder-scoped allow-list that references a `GcpFolder`, start from
**Folder Allow Internal Goto Next**; for blocking by threat list and
country, from **Threat Intel And Geo Block**.
