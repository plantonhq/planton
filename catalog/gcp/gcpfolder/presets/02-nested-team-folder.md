# Nested Team Folder

A folder inside another folder, by reference: `production/payments`. The
chart deploys the parent first, then this one, and every `GcpProject`
whose `folderId` points here inherits both folders' policies and grants.

## What it configures

- `parent.folderId` — a reference to the parent `GcpFolder`'s `folder_id`
  output, so the tree builds in dependency order on a fresh organization.
- `deletionProtection` — left at its default (true).

## Adjust before deploying

- **`parent.folderId.valueFrom.name`** — the parent folder's manifest name.
- **`displayName`** — unique among the parent's children.

## When to choose something else

A top-level folder takes the **Environment Folder** preset.
