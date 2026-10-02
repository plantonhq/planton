# Folder Accepted Risk

## Use Case

Mute low-severity findings across every project in a development folder, keeping live triage for what matters.

## When to Use

- Development folders reviewed on a schedule instead of live
- Reducing alert volume while keeping findings on record

## What This Creates

- A dynamic mute rule for low-severity findings on the folder

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scope.folderId` | `456789012345` | Reference the folder's `GcpFolder`. |
| `filter` | `severity = "LOW"` | Narrow to categories. |
