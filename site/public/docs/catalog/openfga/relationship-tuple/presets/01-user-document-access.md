---
title: "User-Document Access Tuple"
description: "This preset grants a specific user direct access to a document. This is the most fundamental relationship tuple in OpenFGA -- a single user gaining a specific permission on a single resource. It..."
type: "preset"
rank: "01"
presetSlug: "01-user-document-access"
componentSlug: "relationship-tuple"
componentTitle: "Relationship Tuple"
provider: "openfga"
icon: "package"
order: 1
---

# User-Document Access Tuple

This preset grants a specific user direct access to a document. This is the most fundamental relationship tuple in OpenFGA -- a single user gaining a specific permission on a single resource. It corresponds to the OpenFGA format `user:ID` -> `relation` -> `document:ID`.

## When to Use

- Granting individual users access to specific resources
- Direct permission assignments (as opposed to group-based access)
- Any scenario matching the pattern "user X can Y on resource Z"

## Key Configuration Choices

- **Structured user/object** -- uses the proto-correct nested message format with `type` + `id` fields rather than flat strings, enabling cross-resource references via `valueFrom`
- **viewer relation** (`relation: viewer`) -- the most common permission; change to `editor` or `owner` as needed
- **No condition** -- unconditional access; add a `condition` block for time-based or context-based restrictions
- **No model pinning** (`authorizationModelId` omitted) -- uses the latest model version in the store

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<store-id>` | ID of the target OpenFgaStore | `OpenFgaStore` status outputs |
| `<user-id>` | User identifier (e.g., `anne`, `user-123`) | Your identity system |
| `<document-id>` | Document identifier (e.g., `budget-2024`) | Your application |

## Related Presets

- **02-group-membership** -- Use instead when granting access through group membership rather than direct user assignment
