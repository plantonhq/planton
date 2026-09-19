# Project-Scoped Key

A key owned by one project, with a dynamic value pattern: bindings may
carry any `CC-nnnn` code without a declared `GcpTagValue` for each. The
shape that needs only `roles/resourcemanager.tagAdmin` on the project.

## What it configures

- `parent.projectId` — a reference to the owning `GcpProject`; values are
  bindable only inside that project.
- `allowedValuesRegex` — every value must match; the key is DYNAMIC.

## Adjust before deploying

- **`parent.projectId`** — your project, by reference or literal ID.
- **`allowedValuesRegex`** — your code format, or drop it for a fixed set of
  declared values.

## When to choose something else

An estate-wide vocabulary takes the **Environment Key (Organization)** preset.
