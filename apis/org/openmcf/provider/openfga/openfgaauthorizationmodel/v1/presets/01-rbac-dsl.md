# RBAC Authorization Model (DSL)

This preset creates an OpenFGA authorization model using the DSL format that implements role-based access control (RBAC) with user types, groups, and document permissions. This is the most common starting pattern for applications that need viewer/editor/owner roles with group-based access.

## When to Use

- Applications that need role-based access control (viewer, editor, owner)
- Systems where users can be organized into groups with inherited permissions
- Any project starting with OpenFGA that needs a proven, well-understood auth model

## Key Configuration Choices

- **DSL format** (`modelDsl`) -- human-readable, recommended by OpenFGA documentation; the Terraform module auto-converts to JSON
- **Three types** -- `user` (identity), `group` (with members), `document` (with viewer/editor/owner roles)
- **Group inheritance** -- `group#member` can be assigned viewer or editor on documents, granting access to all group members
- **Immutable** -- changing the model creates a new version (new ID); existing tuples remain valid

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<store-id>` | ID of the target OpenFgaStore | `OpenFgaStore` status outputs |

## Related Presets

- **02-document-access-dsl** -- Use instead for a hierarchical document model with folders and inherited permissions
